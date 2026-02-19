package handlers

import (
	"bytes"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/parser"
)

func (h *Handlers) Parse(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"warnings": warnings,
	})
}
