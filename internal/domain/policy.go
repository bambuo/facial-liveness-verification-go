package domain

import (
	"encoding/json"
	"fmt"
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
func ApplyPolicy(base Policy, raw []byte) (Policy, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "{}" {
		return validatePolicy(clampPolicy(base))
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return Policy{}, fmt.Errorf("租户策略不是合法 JSON: %w", err)
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
	return validatePolicy(clampPolicy(base))
}

// validatePolicy 拒绝越界阈值与不可能的帧数组合。
func validatePolicy(p Policy) (Policy, error) {
	check := func(key string, v, min, max float64) error {
		if v < min || v > max {
			return fmt.Errorf("%s 必须在 [%g, %g]，当前 %g", key, min, max, v)
		}
		return nil
	}
	for _, item := range []struct {
		key      string
		value    float64
		min, max float64
	}{
		{"match_threshold", p.MatchThreshold, 0, 1},
		{"liveness_threshold", p.LivenessThreshold, 0, 1},
		{"live_ratio_threshold", p.LiveRatioThreshold, 0, 1},
		{"action_pass_ratio", p.ActionPassRatio, 0, 1},
	} {
		if err := check(item.key, item.value, item.min, item.max); err != nil {
			return Policy{}, err
		}
	}
	if p.YawThreshold <= 0 || p.StraightYaw <= 0 || p.MouthOpenRatio <= 0 {
		return Policy{}, fmt.Errorf("yaw_threshold、straight_yaw、mouth_open_ratio 必须大于 0")
	}
	if p.MaxRequestsPerMinute <= 0 {
		return Policy{}, fmt.Errorf("max_requests_per_minute 必须大于 0")
	}
	actions := 1
	if p.ActionVerifyEnabled {
		actions = p.ActionsCount
	}
	if p.FramesPerAction*actions > 15 {
		return Policy{}, fmt.Errorf("frames_per_action × 动作数不能超过 15，当前 %d × %d", p.FramesPerAction, actions)
	}
	return p, nil
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
