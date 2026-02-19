package handlers

import (
	"bytes"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/enricher"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/parser"
)

func (h *Handlers) Enrich(w http.ResponseWriter, r *http.Request) {
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

	ed := enricher.Enrich(doc)

	if r.URL.Query().Get("generateIds") == "true" {
		enricher.GenerateIDs(ed)
	}

	ed.Document = nil

	writeJSON(w, http.StatusOK, map[string]any{
		"enriched": ed,
		"stats":    ed.Stats,
	})
}
