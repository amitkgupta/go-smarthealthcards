package webformhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/amitkgupta/go-smarthealthcards/ecdsa"
	"github.com/amitkgupta/go-smarthealthcards/fhirbundle"
	"github.com/amitkgupta/go-smarthealthcards/jws"
	"github.com/amitkgupta/go-smarthealthcards/qrcode"
)

type webFormHandler struct {
	key    ecdsa.Key
	issuer string
}

func New(key ecdsa.Key, issuer string) webFormHandler {
	return webFormHandler{key: key, issuer: issuer}
}

func (h webFormHandler) Process(w http.ResponseWriter, r *http.Request) (int, string, bool) {
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

	qrPNG, err := qrcode.Encode(healthCardJWS)
	if err != nil {
		if errors.Is(err, qrcode.JWSTooLargeError) {
			return http.StatusRequestEntityTooLarge, "Breaking up large input into multiple chunks and generating multiple QR codes is not supported at this time.", false
		} else {
			return http.StatusInternalServerError, "", false
		}
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(qrPNG)
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
