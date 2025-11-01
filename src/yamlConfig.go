package src

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config представляет корневую структуру файла wrkit.yaml
type Config struct {
	Vars  map[string]string      `yaml:"vars,omitempty"`
	Tasks map[string]*TaskConfig `yaml:"tasks"`
}

// TaskConfig описывает одну задачу
type TaskConfig struct {
	Desc     string            `yaml:"desc,omitempty"`
	Cmds     StringSlice       `yaml:"cmds"` // поддерживает либо sequence, либо block scalar
	Deps     []string          `yaml:"deps,omitempty"`
	Dir      string            `yaml:"dir,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	Parallel bool              `yaml:"parallel,omitempty"`
}

// StringSlice — кастомный тип для парсинга поля cmds, который принимает:
// - YAML sequence of strings
// - YAML multiline scalar (| или >), который будет разбит по строкам
type StringSlice []string

// UnmarshalYAML реализует кастомную десериализацию для StringSlice
func (s *StringSlice) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.SequenceNode:
		// стандартный список строк
		var arr []string
		if err := node.Decode(&arr); err != nil {
			return fmt.Errorf("decode cmds sequence: %w", err)
		}
		*s = normalizeLines(arr)
		return nil
	case yaml.ScalarNode:
		// мультистрочная строка (block scalar) или одиночная строка
		var raw string
		if err := node.Decode(&raw); err != nil {
			return fmt.Errorf("decode cmds scalar: %w", err)
		}
		// Разбиваем по строкам, убираем пустые и тримим
		lines := splitAndClean(raw)
		*s = normalizeLines(lines)
		return nil
	default:
		return fmt.Errorf("unsupported YAML node kind for cmds: %v", node.Kind)
	}
}

// normalizeLines: убирает пустые строки и тримит, возвращает []string
func normalizeLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		// убрать возможные \r (windows)
		l = strings.ReplaceAll(l, "\r", "")
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}

// splitAndClean: разбивает многострочную строку на строки
func splitAndClean(raw string) []string {
	if raw == "" {
		return nil
	}
	// split by newline, сохранить порядок
	parts := strings.Split(raw, "\n")
	return parts
}

// LoadConfig читает и парсит YAML-конфиг
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if cfg.Tasks == nil {
		cfg.Tasks = map[string]*TaskConfig{}
	}
	if cfg.Vars == nil {
		cfg.Vars = map[string]string{}
	}
	return &cfg, nil
}
