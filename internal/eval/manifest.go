package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest 描述评测集：身份分组、同人图片与活体攻击序列。
type Manifest struct {
	Identities []Identity `yaml:"identities" json:"identities"`
	Liveness   *Liveness  `yaml:"liveness,omitempty" json:"liveness,omitempty"`
	Options    Options    `yaml:"options,omitempty" json:"options,omitempty"`

	Dir  string `yaml:"-" json:"-"`
	Path string `yaml:"-" json:"-"`
}

// Identity 是一个人的图片集合。
type Identity struct {
	ID     string            `yaml:"id" json:"id"`
	Images []string          `yaml:"images" json:"images"`
	Meta   map[string]string `yaml:"meta,omitempty" json:"meta,omitempty"`
}

// Liveness 是带标签的活体序列。
type Liveness struct {
	BonaFide []Sequence `yaml:"bona_fide" json:"bona_fide"`
	Attacks  []Sequence `yaml:"attacks" json:"attacks"`
}

// Sequence 是一组连续帧，label 用于区分攻击类型。
type Sequence struct {
	ID     string   `yaml:"id" json:"id"`
	Label  string   `yaml:"label,omitempty" json:"label,omitempty"`
	Images []string `yaml:"images" json:"images"`
}

// Options 是评测运行参数。
type Options struct {
	MaxImpostorPairs int     `yaml:"max_impostor_pairs,omitempty" json:"max_impostor_pairs,omitempty"`
	ThresholdStep    float64 `yaml:"threshold_step,omitempty" json:"threshold_step,omitempty"`
}

// LoadManifest 读取 YAML/JSON 清单并做基本校验。
func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("解析清单: %w", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	m.Dir = filepath.Dir(abs)
	m.Path = abs
	if m.Options.ThresholdStep <= 0 || m.Options.ThresholdStep > 1 {
		m.Options.ThresholdStep = 0.01
	}
	if len(m.Identities) == 0 {
		return nil, fmt.Errorf("清单至少需要一个 identity")
	}
	seen := make(map[string]bool, len(m.Identities))
	for i := range m.Identities {
		id := &m.Identities[i]
		id.ID = strings.TrimSpace(id.ID)
		if id.ID == "" {
			return nil, fmt.Errorf("identities[%d].id 不能为空", i)
		}
		if seen[id.ID] {
			return nil, fmt.Errorf("identity id 重复: %s", id.ID)
		}
		seen[id.ID] = true
		if len(id.Images) == 0 {
			return nil, fmt.Errorf("identity %s 没有图片", id.ID)
		}
	}
	if m.Liveness != nil {
		if len(m.Liveness.BonaFide) == 0 || len(m.Liveness.Attacks) == 0 {
			return nil, fmt.Errorf("liveness 需要同时提供 bona_fide 与 attacks")
		}
		for i := range m.Liveness.BonaFide {
			if err := validateSequence(&m.Liveness.BonaFide[i], i, "bona_fide"); err != nil {
				return nil, err
			}
		}
		for i := range m.Liveness.Attacks {
			if err := validateSequence(&m.Liveness.Attacks[i], i, "attacks"); err != nil {
				return nil, err
			}
		}
	}
	return &m, nil
}

func validateSequence(s *Sequence, index int, group string) error {
	s.ID = strings.TrimSpace(s.ID)
	if s.ID == "" {
		s.ID = fmt.Sprintf("%s_%d", group, index)
	}
	if len(s.Images) == 0 {
		return fmt.Errorf("%s[%d] 没有图片", group, index)
	}
	if group == "attacks" && strings.TrimSpace(s.Label) == "" {
		s.Label = "unknown"
	}
	return nil
}

// Resolve 将清单中的相对路径解析为相对清单目录的路径。
func (m *Manifest) Resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(m.Dir, path)
}
