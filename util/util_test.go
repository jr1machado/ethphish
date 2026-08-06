package util

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"

	"github.com/gophish/gophish/models"
)

func buildCSVRequest(csvPayload string) (*http.Request, error) {
	csvHeader := "First Name,Last Name,Email\n"
	return buildCSVRequestWithHeader(csvHeader, csvPayload)
}

func buildCSVRequestWithHeader(csvHeader, csvPayload string) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files[]", "example.csv")
	if err != nil {
		return nil, err
	}
	part.Write([]byte(csvHeader))
	part.Write([]byte(csvPayload))
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("POST", "http://127.0.0.1", body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r, nil
}

func TestParseCSVEmail(t *testing.T) {
	expected := models.Target{
		BaseRecipient: models.BaseRecipient{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
		},
	}

	csvPayload := fmt.Sprintf("%s,%s,<%s>", expected.FirstName, expected.LastName, expected.Email)
	r, err := buildCSVRequest(csvPayload)
	if err != nil {
		t.Fatalf("error building CSV request: %v", err)
	}

	got, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("error parsing CSV: %v", err)
	}
	expectedLength := 1
	if len(got) != expectedLength {
		t.Fatalf("invalid number of results received from CSV. expected %d got %d", expectedLength, len(got))
	}
	if !reflect.DeepEqual(expected, got[0]) {
		t.Fatalf("Incorrect targets received. Expected: %#v\nGot: %#v", expected, got)
	}
}

func TestParseCSVProfileFields(t *testing.T) {
	expected := models.Target{
		BaseRecipient: models.BaseRecipient{
			FirstName:  "Ada",
			LastName:   "Lovelace",
			Email:      "ada@example.com",
			Phone:      "+15551234567",
			Position:   "Engineer",
			Custom:     "note",
			Department: "Engineering",
			Company:    "Acme Corp",
			City:       "London",
			State:      "England",
			Country:    "UK",
			Unit:       "Platform",
			Tags:       "vip,executive",
		},
	}

	header := "First Name,Last Name,Email,Phone,Position,Custom,Department,Company,City,State,Country,Unit,Tags\n"
	payload := fmt.Sprintf(
		"%s,%s,<%s>,%s,%s,%s,%s,%s,%s,%s,%s,%s,\"%s\"",
		expected.FirstName, expected.LastName, expected.Email, expected.Phone,
		expected.Position, expected.Custom, expected.Department, expected.Company,
		expected.City, expected.State, expected.Country, expected.Unit, expected.Tags,
	)
	r, err := buildCSVRequestWithHeader(header, payload)
	if err != nil {
		t.Fatalf("error building CSV request: %v", err)
	}

	got, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("error parsing CSV: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("invalid number of results received from CSV. expected 1 got %d", len(got))
	}
	if !reflect.DeepEqual(expected, got[0]) {
		t.Fatalf("Incorrect targets received. Expected: %#v\nGot: %#v", expected, got[0])
	}
}

func TestParseCSVProfileFieldsOptional(t *testing.T) {
	r, err := buildCSVRequest("Ada,Lovelace,<ada@example.com>")
	if err != nil {
		t.Fatalf("error building CSV request: %v", err)
	}

	got, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("error parsing CSV: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("invalid number of results received from CSV. expected 1 got %d", len(got))
	}
	if got[0].Department != "" || got[0].Company != "" || got[0].Tags != "" {
		t.Fatalf("expected empty optional profile fields, got %+v", got[0])
	}
}
