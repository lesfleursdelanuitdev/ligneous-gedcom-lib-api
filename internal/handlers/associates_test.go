package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnrichIncludesAssociates(t *testing.T) {
	h := New()
	req := newGedcomUploadRequest(t, "/api/v1/enrich", `0 HEAD
1 GEDC
2 VERS 5.5
0 @I1@ INDI
1 NAME John /Doe/
1 ASSO @I2@
2 RELA Neighbor
0 @I2@ INDI
1 NAME Jane /Smith/
0 TRLR
`)
	rr := httptest.NewRecorder()
	h.Enrich(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	enriched, ok := body["enriched"].(map[string]any)
	if !ok {
		t.Fatalf("missing enriched object: %+v", body)
	}
	associates, ok := enriched["Associates"].([]any)
	if !ok || len(associates) != 1 {
		t.Fatalf("expected 1 associate edge, got: %#v", enriched["Associates"])
	}
}

func TestPipelineIncludesAssociates(t *testing.T) {
	h := New()
	req := newGedcomUploadRequest(t, "/api/v1/parse-validate-enrich", `0 HEAD
1 GEDC
2 VERS 5.5
0 @I1@ INDI
1 NAME John /Doe/
1 ASSO @I2@
2 RELA Witness
0 @I2@ INDI
1 NAME Jane /Smith/
0 TRLR
`)
	rr := httptest.NewRecorder()
	h.Pipeline(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	enriched, ok := body["enriched"].(map[string]any)
	if !ok {
		t.Fatalf("missing enriched object: %+v", body)
	}
	associates, ok := enriched["Associates"].([]any)
	if !ok || len(associates) != 1 {
		t.Fatalf("expected 1 associate edge, got: %#v", enriched["Associates"])
	}
}

func TestExportGEDCOMWritesAssociationLines(t *testing.T) {
	h := New()
	payload := `{
  "format":"gedcom",
  "filename":"test-asso",
  "enriched":{
    "Individuals":[
      {"xref":"@I1@","full_name":"John /Doe/","full_name_lower":"john /doe/"},
      {"xref":"@I2@","full_name":"Jane /Smith/","full_name_lower":"jane /smith/"}
    ],
    "Associates":[
      {"owner_xref":"@I1@","owner_type":"INDI","associate_xref":"@I2@","relationship":"Neighbor","source_tag":"ASSO"}
    ]
  }
}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/export", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Export(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	out := rr.Body.String()
	if !strings.Contains(out, "1 ASSO @I2@") || !strings.Contains(out, "2 RELA Neighbor") {
		t.Fatalf("expected ASSO/RELA in GEDCOM output, got:\n%s", out)
	}
}

func newGedcomUploadRequest(t *testing.T, path, gedcomText string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.ged")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(gedcomText)); err != nil {
		t.Fatalf("write GEDCOM payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
