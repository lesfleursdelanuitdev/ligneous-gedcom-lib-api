package handlers

import (
	"bytes"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/validator"
)

func (h *Handlers) Validate(w http.ResponseWriter, r *http.Request) {
	data, err := extractGedcomFile(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	doc, _, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	valErrors := validator.Validate(doc)

	errorCount, warnCount, hintCount := 0, 0, 0
	for _, ve := range valErrors {
		switch ve.Severity {
		case validator.SeverityError:
			errorCount++
		case validator.SeverityWarning:
			warnCount++
		case validator.SeverityHint:
			hintCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid":  errorCount == 0,
		"errors": valErrors,
		"counts": map[string]int{
			"errors":   errorCount,
			"warnings": warnCount,
			"hints":    hintCount,
		},
	})
}
