package handlers

import (
	"bytes"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/validator"
)

func (h *Handlers) ParseValidate(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"warnings": warnings,
		"validation": map[string]any{
			"valid":  errorCount == 0,
			"errors": valErrors,
		},
	})
}
