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

type FHIRBundle struct {
	Patient
	Immunizations []Immunization
}

type Patient struct {
	Name
	BirthDate time.Time
}

type Name struct {
	Family string   `json:"family"`
	Givens []string `json:"given"`
}

type Immunization struct {
	DatePerformed time.Time
	Performer     string
	LotNumber     string
	VaccineType
}

type VaccineType string

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
