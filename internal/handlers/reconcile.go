package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/enricher"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/reconciliation"
)

const maxReconcileJSONBytes = 25 << 20

type reconcileRequestBody struct {
	Left    json.RawMessage `json:"left"`
	Right   json.RawMessage `json:"right"`
	Options json.RawMessage `json:"options"`
}

func (h *Handlers) parseReconcileBody(body []byte) (left *enricher.EnrichedDocument, right *enricher.EnrichedDocument, opt *reconciliation.Options, errMsg string, status int) {
	var req reconcileRequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, nil, nil, "json: " + err.Error(), http.StatusBadRequest
	}
	if len(req.Left) == 0 || len(req.Right) == 0 {
		return nil, nil, nil, "left and right enriched documents are required", http.StatusBadRequest
	}
	opt, err := reconciliation.MergeReconcileOptionsFromJSON(req.Options)
	if err != nil {
		return nil, nil, nil, "options: " + err.Error(), http.StatusBadRequest
	}
	left, err = reconciliation.DecodeEnrichedDocument(req.Left)
	if err != nil {
		return nil, nil, nil, "left: " + err.Error(), http.StatusUnprocessableEntity
	}
	right, err = reconciliation.DecodeEnrichedDocument(req.Right)
	if err != nil {
		return nil, nil, nil, "right: " + err.Error(), http.StatusUnprocessableEntity
	}
	return left, right, opt, "", 0
}

// ReconcileMergePlan POST JSON body: { "left": EnrichedDocument, "right": EnrichedDocument, "options"?: { ... } }
// Returns a declarative MergePlan (no DB writes, no mutation).
// Options are merged onto package defaults (soft matching on unless explicitly disabled).
func (h *Handlers) ReconcileMergePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxReconcileJSONBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	if len(body) > maxReconcileJSONBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "body exceeds limit")
		return
	}
	left, right, opt, errMsg, status := h.parseReconcileBody(body)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	plan := reconciliation.BuildMergePlan(left, right, opt)
	writeJSON(w, http.StatusOK, map[string]any{"mergePlan": plan})
}

// ReconcileSession POST same body as ReconcileMergePlan; returns mergePlan plus a draft ReconciliationSession (for persistence / UI).
func (h *Handlers) ReconcileSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxReconcileJSONBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	if len(body) > maxReconcileJSONBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "body exceeds limit")
		return
	}
	left, right, opt, errMsg, status := h.parseReconcileBody(body)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	plan := reconciliation.BuildMergePlan(left, right, opt)
	sess := reconciliation.NewReconciliationSession(plan)
	sess.WithInputSummary(len(left.Individuals), len(right.Individuals))
	writeJSON(w, http.StatusOK, map[string]any{
		"mergePlan": plan,
		"session":   sess,
	})
}
