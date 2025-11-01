package src

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config описывает структуру wrkit.yaml
type Config struct {
	Vars  map[string]string      `yaml:"vars,omitempty"`
	Tasks map[string]*TaskConfig `yaml:"tasks"`
}

// TaskConfig — описание одной задачи
type TaskConfig struct {
	Desc     string            `yaml:"desc,omitempty"`
	Cmds     StringSlice       `yaml:"cmds"`
	Deps     []string          `yaml:"deps,omitempty"`
	Dir      string            `yaml:"dir,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	Parallel bool              `yaml:"parallel,omitempty"`
}

// StringSlice поддерживает YAML sequence или block scalar
type StringSlice []string

func (s *StringSlice) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.SequenceNode:
		var arr []string
		if err := node.Decode(&arr); err != nil {
			return fmt.Errorf("decode cmds sequence: %w", err)
		}
		*s = normalizeLines(arr)
		return nil
	case yaml.ScalarNode:
		var raw string
		if err := node.Decode(&raw); err != nil {
			return fmt.Errorf("decode cmds scalar: %w", err)
		}
		lines := splitAndClean(raw)
		*s = normalizeLines(lines)
		return nil
	default:
		return fmt.Errorf("unsupported YAML node kind for cmds: %v", node.Kind)
	}
}

func normalizeLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(strings.ReplaceAll(l, "\r", ""))
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}

func splitAndClean(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

// LoadConfig читает YAML конфиг, возвращает (nil, nil) если файл отсутствует
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // нет файла — не ошибка
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml %s: %w", path, err)
	}
	if cfg.Tasks == nil {
		cfg.Tasks = map[string]*TaskConfig{}
	}
	if cfg.Vars == nil {
		cfg.Vars = map[string]string{}
	}
	return &cfg, nil
}

// MergeVars объединяет переменные из конфига, окружения и CLI
func MergeVars(cfg *Config, cliVars map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range cfg.Vars {
		merged[k] = v
	}

	envMap := map[string]string{}
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for k, v := range envMap {
		merged["env."+k] = v // добавляем с префиксом
	}

	for k, v := range cliVars {
		merged[k] = v
	}

	return merged
}

// LoadCombinedConfig ищет локальный wrkit.yaml и глобальный ~/.wrkit.master.yaml
// если noMaster == true, используется только локальный файл
func LoadCombinedConfig(localPath string, noMaster bool) (*Config, error) {
	var masterCfg *Config
	var err error

	localCfg, err := LoadConfig(localPath)
	if err != nil {
		return nil, err
	}

	if !noMaster {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		masterPath := filepath.Join(homeDir, ".wrkit.master.yaml")

		masterCfg, err = LoadConfig(masterPath)
		if err != nil {
			return nil, err
		}
	}

	// если нет ни одного файла — вернём пустой конфиг
	if localCfg == nil && masterCfg == nil {
		return &Config{Vars: map[string]string{}, Tasks: map[string]*TaskConfig{}}, nil
	}
	if localCfg == nil {
		return masterCfg, nil
	}
	if masterCfg == nil || noMaster {
		return localCfg, nil
	}

	// Объединяем: приоритет у локального
	merged := &Config{
		Vars:  map[string]string{},
		Tasks: map[string]*TaskConfig{},
	}

	for k, v := range masterCfg.Vars {
		merged.Vars[k] = v
	}
	for k, v := range localCfg.Vars {
		merged.Vars[k] = v
	}

	for k, t := range masterCfg.Tasks {
		merged.Tasks[k] = t
	}
	for k, t := range localCfg.Tasks {
		merged.Tasks[k] = t
	}

	return merged, nil
}
