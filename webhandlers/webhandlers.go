// Package webhandlers can be used in a web-based application for issuing SMART
// Health Card QR codes for COVID-19 immunizations.
package webhandlers

import (
	"archive/zip"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amitkgupta/go-smarthealthcards/v2/fhirbundle"
	"github.com/amitkgupta/go-smarthealthcards/v2/jws"
	"github.com/amitkgupta/go-smarthealthcards/v2/qrcode"
)

// Handlers should not be instantiated directly; use the New
// function in this package instead.
type Handlers struct {
	key    *ecdsa.PrivateKey
	issuer string
}

// New returns an object with methods that can be used in a web-based
// application for issuing SMART Health Card QR codes for COVID-19
// immunizations.
func New(key *ecdsa.PrivateKey, issuer string) Handlers {
	return Handlers{key: key, issuer: issuer}
}

// JWKSJSON writes the JSON representation of the JSON Web Key Set
// representation of the public information of the associated private
// key.
//
// If there is an error, this methods returns the HTTP response code,
// an additional error message if available, and false. If there is no
// error, it returns 0, the empty string, and true.
func (h Handlers) JWKSJSON(w http.ResponseWriter) (int, string, bool) {
	if jwksJSON, err := jws.JWKSJSON(h.key); err != nil {
		return http.StatusInternalServerError, "", false
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksJSON)
		return 0, "", true
	}
}

// ProcessForm expects the request to provide form data representing a patient
// and his or her COVID-19 immunizations. This method extracts the form values
// from the request, constructs an FHIR bundle from the form data, creates and
// signs a JSON Web Signature encapsulating that data, and either writes a PNG
// image of a single QR code representing a SMART Health Card with the
// immunziation data, or a ZIP archive consisting of multiple PNGs of QR codes
// which can be combined into a single SMART Health Card with the immunization
// data.
//
// If there is an error, this methods returns the HTTP response code,
// an additional error message if available, and false. If there is no
// error, it returns 0, the empty string, and true.
func (h Handlers) ProcessForm(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	fhirBundle, err := parseInput(r)
	if err != nil {
		return http.StatusBadRequest, err.Error(), false
	}

	payload, err := json.Marshal(fhirbundle.NewJWSPayload(fhirBundle, h.issuer))
	if err != nil {
		return http.StatusInternalServerError, "", false
	}

	healthCardJWS, err := jws.SignAndSerialize(payload, h.key)
	if err != nil {
		return http.StatusInternalServerError, "", false
	}

	qrPNGs, err := qrcode.Encode(healthCardJWS)
	if err != nil {
		return http.StatusInternalServerError, "", false
	}

	if len(qrPNGs) == 1 {
		w.Header().Set("Content-Type", "image/png")
		w.Write(qrPNGs[0])
	} else {
		zw := zip.NewWriter(w)

		for i, qrPNG := range qrPNGs {
			if f, err := zw.Create(fmt.Sprintf("%d.png", i+1)); err != nil {
				return http.StatusInternalServerError, "", false
			} else if _, err = f.Write(qrPNG); err != nil {
				return http.StatusInternalServerError, "", false
			}
		}

		if err := zw.Close(); err != nil {
			return http.StatusInternalServerError, "", false
		}

		w.Header().Set("Content-Type", "application/zip")
	}

	return 0, "", true
}

func parseInput(r *http.Request) (fhirbundle.FHIRBundle, error) {
	familyName := strings.TrimSpace(r.PostFormValue("family_name"))
	givenNames := strings.TrimSpace(r.PostFormValue("given_names"))
	birthDateString := strings.TrimSpace(r.PostFormValue("date_of_birth"))

	firstImmunizationPerformer := strings.TrimSpace(r.PostFormValue("first_immunization_performer"))
	firstImmunizationLotNumber := strings.TrimSpace(r.PostFormValue("first_immunization_lot_number"))
	firstImmunizationVaccineTypeString := strings.TrimSpace(r.PostFormValue("first_immunization_vaccine_type"))
	firstImmunizationDateString := strings.TrimSpace(r.PostFormValue("first_immunization_date"))

	secondImmunizationPerformer := strings.TrimSpace(r.PostFormValue("second_immunization_performer"))
	secondImmunizationLotNumber := strings.TrimSpace(r.PostFormValue("second_immunization_lot_number"))
	secondImmunizationVaccineTypeString := strings.TrimSpace(r.PostFormValue("second_immunization_vaccine_type"))
	secondImmunizationDateString := strings.TrimSpace(r.PostFormValue("second_immunization_date"))

	thirdImmunizationPerformer := strings.TrimSpace(r.PostFormValue("third_immunization_performer"))
	thirdImmunizationLotNumber := strings.TrimSpace(r.PostFormValue("third_immunization_lot_number"))
	thirdImmunizationVaccineTypeString := strings.TrimSpace(r.PostFormValue("third_immunization_vaccine_type"))
	thirdImmunizationDateString := strings.TrimSpace(r.PostFormValue("third_immunization_date"))

	if familyName == "" || givenNames == "" || birthDateString == "" ||
		firstImmunizationPerformer == "" || firstImmunizationLotNumber == "" ||
		firstImmunizationVaccineTypeString == "" || firstImmunizationDateString == "" {
		return fhirbundle.FHIRBundle{}, errors.New("patient information or first immunization information missing")
	}

	if (secondImmunizationPerformer != "" || secondImmunizationLotNumber != "" ||
		secondImmunizationVaccineTypeString != "" || secondImmunizationDateString != "") &&
		(secondImmunizationPerformer == "" || secondImmunizationLotNumber == "" ||
			secondImmunizationVaccineTypeString == "" || secondImmunizationDateString == "") {
		return fhirbundle.FHIRBundle{}, errors.New("second immunization information only partially complete")
	}

	if (thirdImmunizationPerformer != "" || thirdImmunizationLotNumber != "" ||
		thirdImmunizationVaccineTypeString != "" || thirdImmunizationDateString != "") &&
		(secondImmunizationPerformer == "") {
		return fhirbundle.FHIRBundle{}, errors.New("third immunization information provided while second immunization is blank")
	}

	if (thirdImmunizationPerformer != "" || thirdImmunizationLotNumber != "" ||
		thirdImmunizationVaccineTypeString != "" || thirdImmunizationDateString != "") &&
		(thirdImmunizationPerformer == "" || thirdImmunizationLotNumber == "" ||
			thirdImmunizationVaccineTypeString == "" || thirdImmunizationDateString == "") {
		return fhirbundle.FHIRBundle{}, errors.New("third immunization information only partially complete")
	}

	birthDate, err := time.Parse("2006-01-02", birthDateString)
	if err != nil {
		return fhirbundle.FHIRBundle{}, errors.New("invalid patient birth date")
	}

	patient := fhirbundle.Patient{
		Name: fhirbundle.Name{
			Family: familyName,
			Givens: strings.Fields(givenNames),
		},
		BirthDate: birthDate,
	}

	firstImmunizationDate, err := time.Parse("2006-01-02", firstImmunizationDateString)
	if err != nil {
		return fhirbundle.FHIRBundle{}, errors.New("invalid first immunization date")
	}

	firstImmunizationVaccineType := fhirbundle.VaccineType(firstImmunizationVaccineTypeString)
	switch firstImmunizationVaccineType {
	case fhirbundle.Pfizer, fhirbundle.Moderna, fhirbundle.JohnsonAndJohnson,
		fhirbundle.AstraZeneca, fhirbundle.Sinopharm, fhirbundle.COVAXIN:
	default:
		return fhirbundle.FHIRBundle{}, errors.New("invalid first immunization vaccine type")
	}

	immunizations := []fhirbundle.Immunization{
		{
			DatePerformed: firstImmunizationDate,
			Performer:     firstImmunizationPerformer,
			LotNumber:     firstImmunizationLotNumber,
			VaccineType:   firstImmunizationVaccineType,
		},
	}

	if secondImmunizationPerformer != "" {
		secondImmunizationDate, err := time.Parse("2006-01-02", secondImmunizationDateString)
		if err != nil {
			return fhirbundle.FHIRBundle{}, errors.New("invalid second immunization date")
		}

		secondImmunizationVaccineType := fhirbundle.VaccineType(secondImmunizationVaccineTypeString)
		switch secondImmunizationVaccineType {
		case fhirbundle.Pfizer, fhirbundle.Moderna, fhirbundle.JohnsonAndJohnson,
			fhirbundle.AstraZeneca, fhirbundle.Sinopharm, fhirbundle.COVAXIN:
		default:
			return fhirbundle.FHIRBundle{}, errors.New("invalid second immunization vaccine type")
		}

		immunizations = append(immunizations, fhirbundle.Immunization{
			DatePerformed: secondImmunizationDate,
			Performer:     secondImmunizationPerformer,
			LotNumber:     secondImmunizationLotNumber,
			VaccineType:   secondImmunizationVaccineType,
		})
	}

	if thirdImmunizationPerformer != "" {
		thirdImmunizationDate, err := time.Parse("2006-01-02", thirdImmunizationDateString)
		if err != nil {
			return fhirbundle.FHIRBundle{}, errors.New("invalid third immunization date")
		}

		thirdImmunizationVaccineType := fhirbundle.VaccineType(thirdImmunizationVaccineTypeString)
		switch thirdImmunizationVaccineType {
		case fhirbundle.Pfizer, fhirbundle.Moderna, fhirbundle.JohnsonAndJohnson,
			fhirbundle.AstraZeneca, fhirbundle.Sinopharm, fhirbundle.COVAXIN:
		default:
			return fhirbundle.FHIRBundle{}, errors.New("invalid third immunization vaccine type")
		}

		immunizations = append(immunizations, fhirbundle.Immunization{
			DatePerformed: thirdImmunizationDate,
			Performer:     thirdImmunizationPerformer,
			LotNumber:     thirdImmunizationLotNumber,
			VaccineType:   thirdImmunizationVaccineType,
		})
	}

	return fhirbundle.FHIRBundle{Patient: patient, Immunizations: immunizations}, nil
}
