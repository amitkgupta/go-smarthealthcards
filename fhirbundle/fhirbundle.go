// Package fhirbundle constructs and marshals a (pre-compressed) JWS
// payload containing an FHIR bundle of information representing
// COVID-19 immunizations. See
// https://spec.smarthealth.cards/#health-cards-are-encoded-as-compact-serialization-json-web-signatures-jws
// and
// https://build.fhir.org/ig/HL7/fhir-shc-vaccination-ig/StructureDefinition-shc-vaccination-bundle-dm.html#tab-snapshot.
package fhirbundle

import (
	"encoding/json"
	"fmt"
	"time"
)

type jwsPayload struct {
	Issuer                string                `json:"iss"`
	NotBefore             int64                 `json:"nbf"`
	VerifiableCredentials verifiableCredentials `json:"vc"`
}

type verifiableCredentials struct {
	Type              []string          `json:"type"`
	CredentialSubject credentialSubject `json:"credentialSubject"`
}

type credentialSubject struct {
	Version string     `json:"fhirVersion"`
	Bundle  FHIRBundle `json:"fhirBundle"`
}

// NewJWSPayload returns a struct that can be serialized as JSON
// and represent the (pre-compressed) payload of a JSON Web Signature
// (JWS) as described here:
// https://spec.smarthealth.cards/#health-cards-are-encoded-as-compact-serialization-json-web-signatures-jws.
//
// This function takes the core relevant data for an FHIR
// bundle representing a patient's COVID-19 immunizations,
// encapsulated in an FHIRBundle object, and an issuer which
// is the entity that will JWS, as inputs.
func NewJWSPayload(fb FHIRBundle, issuer string) jwsPayload {
	return jwsPayload{
		Issuer:    issuer,
		NotBefore: time.Now().Unix(),
		VerifiableCredentials: verifiableCredentials{
			Type: []string{
				"https://smarthealth.cards#health-card",
				"https://smarthealth.cards#immunization",
				"https://smarthealth.cards#covid19",
			},
			CredentialSubject: credentialSubject{
				Version: "4.0.1",
				Bundle:  fb,
			},
		},
	}
}

// FHIRBundle encapsulates the core relevant data for an FHIR
// bundle representing a patient's COVID-19 immunizations.
type FHIRBundle struct {
	// Patient represents an individual who has received immunizations.
	Patient

	// Immunizations represents the immunizations the patient has received.
	Immunizations []Immunization
}

// Patient represents an individual who has received immunizations.
type Patient struct {
	// Name is the patient's name.
	Name

	// BirthDate is the patient's date of birth.
	BirthDate time.Time
}

// Name represents a patient's name.
type Name struct {
	// Family represents the patient's family name.
	Family string `json:"family"`

	// Givens represents the patient's given names.
	Givens []string `json:"given"`
}

// Immunization represents one instance of a COVID-19 immunization
// performed on a patient.
type Immunization struct {
	// DatePerformed represents the date when the immunization was
	// performed.
	DatePerformed time.Time

	// Performer represents the entity which performed the immunization
	// such as a particular hospital or health clinic.
	Performer string

	// LotNumber represents the lot number of the specific batch of the
	// vaccine that was administered.
	LotNumber string

	// VaccineType represents the type of vaccine that was administered,
	// e.g. Pfizer-BioNTech.
	VaccineType
}

type VaccineType string

// Supported COVID-19 vaccination types.
const (
	Pfizer            VaccineType = "Pfizer"
	Moderna           VaccineType = "Moderna"
	JohnsonAndJohnson VaccineType = "JohnsonAndJohnson"
	AstraZeneca       VaccineType = "AstraZeneca"
	Sinopharm         VaccineType = "Sinopharm"
	COVAXIN           VaccineType = "COVAXIN"
)

// https://www2a.cdc.gov/vaccines/iis/iisstandards/vaccines.asp?rpt=cvx
func (vt VaccineType) cvxcode() string {
	switch vt {
	case Pfizer:
		return "208"
	case Moderna:
		return "207"
	case JohnsonAndJohnson:
		return "212"
	case AstraZeneca:
		return "210"
	case Sinopharm:
		return "510"
	case COVAXIN:
		return "502"
	}

	panic("cvxcode called on invalid VaccineType")
}

type fhirBundleJSON struct {
	ResourceType string      `json:"resourceType"`
	Type         string      `json:"type"`
	Entries      []entryJSON `json:"entry"`
}

type entryJSON struct {
	FullURL  string       `json:"fullUrl"`
	Resource resourceJSON `json:"resource"`
}

type resourceJSON struct {
	ResourceType   string           `json:"resourceType"`
	Name           []Name           `json:"name,omitempty"`
	BirthDate      string           `json:"birthDate,omitempty"`
	Status         string           `json:"status,omitempty"`
	VaccineCode    *vaccineCodeJSON `json:"vaccineCode,omitempty"`
	Patient        *patientJSON     `json:"patient,omitempty"`
	OccurrenceDate string           `json:"occurrenceDateTime,omitempty"`
	Performers     []performerJSON  `json:"performer,omitempty"`
	LotNumber      string           `json:"lotNumber,omitempty"`
}

type vaccineCodeJSON struct {
	Coding []codingJSON `json:"coding,omitempty"`
}

type codingJSON struct {
	System string `json:"system"`
	Code   string `json:"code"`
}

type patientJSON struct {
	Reference string `json:"reference,omitempty"`
}

type performerJSON struct {
	Actor actorJSON `json:"actor"`
}

type actorJSON struct {
	Display string `json:"display"`
}

// MarshalJSON takes the core relevant data for an FHIR bundle
// encapsulated in an FHIRBundle object, and seralizes it as
// a JSON byte slice including all the additional boilerplate
// as defined here:
// https://build.fhir.org/ig/HL7/fhir-shc-vaccination-ig/StructureDefinition-shc-vaccination-bundle-dm.html.
func (f FHIRBundle) MarshalJSON() ([]byte, error) {
	fbj := fhirBundleJSON{
		ResourceType: "Bundle",
		Type:         "collection",
		Entries:      make([]entryJSON, len(f.Immunizations)+1),
	}

	fbj.Entries[0] = entryJSON{
		FullURL: "resource:0",
		Resource: resourceJSON{
			ResourceType: "Patient",
			Name:         []Name{f.Patient.Name},
			BirthDate:    f.Patient.BirthDate.Format("2006-01-02"),
		},
	}

	for i, immunization := range f.Immunizations {
		fbj.Entries[i+1] = entryJSON{
			FullURL: fmt.Sprintf("resource:%d", i+1),
			Resource: resourceJSON{
				ResourceType: "Immunization",
				Status:       "completed",
				VaccineCode: &(vaccineCodeJSON{
					Coding: []codingJSON{
						{
							System: "https://hl7.org/fhir/sid/cvx", // https://www.hl7.org/fhir/cvx.html
							Code:   immunization.VaccineType.cvxcode(),
						},
					},
				}),
				Patient:        &(patientJSON{Reference: "resource:0"}),
				OccurrenceDate: immunization.DatePerformed.Format("2006-01-02"),
				Performers:     []performerJSON{{Actor: actorJSON{Display: immunization.Performer}}},
				LotNumber:      immunization.LotNumber,
			},
		}
	}

	return json.Marshal(&fbj)
}
