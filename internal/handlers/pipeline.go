package handlers

import (
	"bytes"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/enricher"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/validator"
)

func (h *Handlers) Pipeline(w http.ResponseWriter, r *http.Request) {
	data, err := extractGedcomFile(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	doc, warnings, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	valErrors := validator.Validate(doc)

	errorCount := 0
	for _, ve := range valErrors {
		if ve.Severity == validator.SeverityError {
			errorCount++
		}
	}

	ed := enricher.Enrich(doc)

	if r.URL.Query().Get("generateIds") == "true" {
		enricher.GenerateIDs(ed)
	}

	// Nil out embedded document since we return it separately at top level
	ed.Document = nil

	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"warnings": warnings,
		"validation": map[string]any{
			"valid":  errorCount == 0,
			"errors": valErrors,
		},
		"enriched": ed,
		"stats":    ed.Stats,
	})
}
