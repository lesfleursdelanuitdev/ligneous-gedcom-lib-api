package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/enricher"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/exporter"
)

func (h *Handlers) Export(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Enriched *enricher.EnrichedDocument `json:"enriched"`
		Format   string                     `json:"format"`
		Filename string                     `json:"filename"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if payload.Enriched == nil {
		writeError(w, http.StatusBadRequest, "missing 'enriched' field")
		return
	}

	format := payload.Format
	if format == "" {
		format = "gedcom"
	}
	filename := payload.Filename
	if filename == "" {
		filename = "export"
	}

	switch format {
	case "gedcom":
		// Always reconstruct from enriched tables so NOTE xrefs, DATE values, and OBJE
		// match what JSON/CSV use. (WithOriginal would serialize a raw Document when
		// present, which can bypass normalizers and confuse DB-only payloads.)
		gedcomText := exporter.EnrichedToGEDCOM(payload.Enriched)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".ged\"")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(gedcomText))

	case "json":
		data := exporter.FromEnrichedToJSON(payload.Enriched)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".json\"")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(data)

	case "csv":
		csvText, err := exporter.FromEnrichedToCSV(payload.Enriched)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "csv export: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".csv\"")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(csvText))

	default:
		writeError(w, http.StatusBadRequest, "unsupported format: "+format+". Use 'gedcom', 'json', or 'csv'")
	}
}
