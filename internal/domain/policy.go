package domain

import (
	"encoding/json"
	"math/rand"
	"strings"
)

// Policy 是按租户合并后的生效策略。
type Policy struct {
	MatchThreshold       float64
	LivenessThreshold    float64
	LiveRatioThreshold   float64
	YawThreshold         float64
	StraightYaw          float64
	MouthOpenRatio       float64
	ActionPassRatio      float64
	FramesPerAction      int
	ActionsCount         int
	ActionVerifyEnabled  bool
	MaxRequestsPerMinute int
}

// BuildActions 生成动作序列：straight 打底，其余动作洗牌后取前 N 个。
func (p Policy) BuildActions(r *rand.Rand) []string {
	actions := []string{"straight"}
	if !p.ActionVerifyEnabled || p.ActionsCount <= 1 {
		return actions
	}
	pool := []string{"turn_left", "turn_right", "mouth_open"}
	for i := len(pool) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		pool[i], pool[j] = pool[j], pool[i]
	}
	n := p.ActionsCount - 1
	if n > len(pool) {
		n = len(pool)
	}
	return append(actions, pool[:n]...)
}

// ApplyPolicy 以全局策略为基线，用租户 policy_json 覆盖已知键。
func ApplyPolicy(base Policy, raw []byte) Policy {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "{}" {
		return clampPolicy(base)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return clampPolicy(base)
	}
	if v := getFloat(m, "match_threshold"); v != nil {
		base.MatchThreshold = *v
	}
	if v := getFloat(m, "liveness_threshold"); v != nil {
		base.LivenessThreshold = *v
	}
	if v := getFloat(m, "live_ratio_threshold"); v != nil {
		base.LiveRatioThreshold = *v
	}
	if v := getFloat(m, "yaw_threshold"); v != nil {
		base.YawThreshold = *v
	}
	if v := getFloat(m, "straight_yaw"); v != nil {
		base.StraightYaw = *v
	}
	if v := getFloat(m, "mouth_open_ratio"); v != nil {
		base.MouthOpenRatio = *v
	}
	if v := getFloat(m, "action_pass_ratio"); v != nil {
		base.ActionPassRatio = *v
	}
	if v := getInt(m, "frames_per_action"); v != nil {
		base.FramesPerAction = *v
	}
	if v := getInt(m, "actions_count"); v != nil {
		base.ActionsCount = *v
	}
	if v := getBool(m, "action_verify_enabled"); v != nil {
		base.ActionVerifyEnabled = *v
	}
	if v := getInt(m, "max_requests_per_minute"); v != nil {
		base.MaxRequestsPerMinute = *v
	}
	return clampPolicy(base)
}

func clampPolicy(p Policy) Policy {
	if p.FramesPerAction < 2 {
		p.FramesPerAction = 2
	}
	if p.FramesPerAction > 8 {
		p.FramesPerAction = 8
	}
	if p.ActionsCount < 1 {
		p.ActionsCount = 1
	}
	if p.ActionsCount > 3 {
		p.ActionsCount = 3
	}
	return p
}

func getFloat(m map[string]json.RawMessage, key string) *float64 {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	var v float64
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return &v
}

func getInt(m map[string]json.RawMessage, key string) *int {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	var v int
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return &v
}

func getBool(m map[string]json.RawMessage, key string) *bool {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	var v bool
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return &v
}
