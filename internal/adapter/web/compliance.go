package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleRulesetCreate 实现 POST /api/v2/compliance/rulesets。
func (s *Server) handleRulesetCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"is_active"`
		Priority    *int32  `json:"priority"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.Name == "" || len(req.Name) > 200 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name is required (max 200 chars)"})
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	priority := int32(100)
	if req.Priority != nil {
		priority = *req.Priority
	}
	now := time.Now().UTC()
	ruleset := &gormmodel.ComplianceRuleset{
		ID: uuid.New(), DeveloperID: principalFrom(r).DeveloperID, Name: req.Name,
		Description: req.Description, IsActive: active, Priority: priority, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.idswyft.CreateRuleset(r.Context(), ruleset); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "ruleset": ruleset})
}

// handleRulesetList 实现 GET /api/v2/compliance/rulesets。
func (s *Server) handleRulesetList(w http.ResponseWriter, r *http.Request) {
	rulesets, err := s.idswyft.ListRulesets(r.Context(), principalFrom(r).DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(rulesets))
	for _, ruleset := range rulesets {
		rules, _ := s.idswyft.RulesByRuleset(r.Context(), ruleset.ID)
		out = append(out, map[string]any{
			"id": ruleset.ID, "name": ruleset.Name, "description": ruleset.Description,
			"is_active": ruleset.IsActive, "priority": ruleset.Priority,
			"created_at": ruleset.CreatedAt, "updated_at": ruleset.UpdatedAt, "rule_count": len(rules),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "rulesets": out})
}

// handleRulesetGet 实现 GET /api/v2/compliance/rulesets/:id。
func (s *Server) handleRulesetGet(w http.ResponseWriter, r *http.Request) {
	ruleset, ok := s.ruleset(w, r)
	if !ok {
		return
	}
	rules, _ := s.idswyft.RulesByRuleset(r.Context(), ruleset.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "ruleset": map[string]any{
		"id": ruleset.ID, "name": ruleset.Name, "description": ruleset.Description,
		"is_active": ruleset.IsActive, "priority": ruleset.Priority,
		"created_at": ruleset.CreatedAt, "updated_at": ruleset.UpdatedAt, "rules": rules,
	}})
}

// handleRulesetUpdate 实现 PUT /api/v2/compliance/rulesets/:id。
func (s *Server) handleRulesetUpdate(w http.ResponseWriter, r *http.Request) {
	ruleset, ok := s.ruleset(w, r)
	if !ok {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"is_active"`
		Priority    *int32  `json:"priority"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if _, err := s.idswyft.UpdateRuleset(r.Context(), principalFrom(r).DeveloperID, ruleset.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	updated, _ := s.idswyft.GetRuleset(r.Context(), principalFrom(r).DeveloperID, ruleset.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "ruleset": updated})
}

// handleRulesetDelete 实现 DELETE /api/v2/compliance/rulesets/:id。
func (s *Server) handleRulesetDelete(w http.ResponseWriter, r *http.Request) {
	ruleset, ok := s.ruleset(w, r)
	if !ok {
		return
	}
	if _, err := s.idswyft.DeleteRuleset(r.Context(), principalFrom(r).DeveloperID, ruleset.ID); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Ruleset and all rules deleted"})
}

// handleRuleCreate 实现 POST /api/v2/compliance/rulesets/:id/rules。
func (s *Server) handleRuleCreate(w http.ResponseWriter, r *http.Request) {
	ruleset, ok := s.ruleset(w, r)
	if !ok {
		return
	}
	var req struct {
		Condition   json.RawMessage `json:"condition"`
		Action      json.RawMessage `json:"action"`
		Description *string         `json:"description"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if len(req.Condition) == 0 || len(req.Action) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "condition and action are required"})
		return
	}
	rule := &gormmodel.ComplianceRule{
		ID: uuid.New(), RulesetID: ruleset.ID, Condition: req.Condition, Action: req.Action,
		Description: req.Description, CreatedAt: time.Now().UTC(),
	}
	if err := s.idswyft.CreateRule(r.Context(), rule); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "rule": rule})
}

// handleRuleUpdate 实现 PUT /api/v2/compliance/rules/:id。
func (s *Server) handleRuleUpdate(w http.ResponseWriter, r *http.Request) {
	rule, ok := s.rule(w, r)
	if !ok {
		return
	}
	var req struct {
		Condition   json.RawMessage `json:"condition"`
		Action      json.RawMessage `json:"action"`
		Description *string         `json:"description"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{}
	if len(req.Condition) > 0 {
		updates["condition"] = req.Condition
	}
	if len(req.Action) > 0 {
		updates["action"] = req.Action
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "No fields to update"})
		return
	}
	if _, err := s.idswyft.UpdateRule(r.Context(), rule.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	updated, _ := s.idswyft.GetRule(r.Context(), rule.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "rule": updated})
}

// handleRuleDelete 实现 DELETE /api/v2/compliance/rules/:id。
func (s *Server) handleRuleDelete(w http.ResponseWriter, r *http.Request) {
	rule, ok := s.rule(w, r)
	if !ok {
		return
	}
	if _, err := s.idswyft.DeleteRule(r.Context(), rule.ID); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Rule deleted"})
}

// handleComplianceEvaluate 实现 POST /api/v2/compliance/evaluate。
func (s *Server) handleComplianceEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Context map[string]any `json:"context"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.Context == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "context object is required"})
		return
	}
	rulesets, err := s.idswyft.ListRulesets(r.Context(), principalFrom(r).DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	matches := make([]map[string]any, 0)
	merged := map[string]any{}
	for _, ruleset := range rulesets {
		if !ruleset.IsActive {
			continue
		}
		rules, _ := s.idswyft.RulesByRuleset(r.Context(), ruleset.ID)
		for _, rule := range rules {
			var condition map[string]any
			_ = json.Unmarshal(rule.Condition, &condition)
			if conditionMatches(condition, req.Context) {
				var action map[string]any
				_ = json.Unmarshal(rule.Action, &action)
				for key, value := range action {
					merged[key] = value
				}
				matches = append(matches, map[string]any{"ruleset": ruleset.Name, "rule": rule.ID.String(), "action": action})
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "matched_rules": len(matches), "matches": matches, "resolved_action": merged})
}

func (s *Server) ruleset(w http.ResponseWriter, r *http.Request) (*gormmodel.ComplianceRuleset, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid ruleset ID"})
		return nil, false
	}
	ruleset, err := s.idswyft.GetRuleset(r.Context(), principalFrom(r).DeveloperID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Ruleset not found"})
		return nil, false
	}
	return ruleset, true
}

func (s *Server) rule(w http.ResponseWriter, r *http.Request) (*gormmodel.ComplianceRule, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid rule ID"})
		return nil, false
	}
	rule, err := s.idswyft.GetRule(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Rule not found"})
		return nil, false
	}
	ruleset, err := s.idswyft.GetRuleset(r.Context(), principalFrom(r).DeveloperID, rule.RulesetID)
	if err != nil || ruleset == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Rule not found"})
		return nil, false
	}
	return rule, true
}

func conditionMatches(condition, context map[string]any) bool {
	for key, expected := range condition {
		if actual, ok := context[key]; !ok || !equalValue(actual, expected) {
			return false
		}
	}
	return true
}

func equalValue(a, b any) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}
