package config

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	kOpenAI   = "openai"
	kRpm      = "rpm"
	kReqArgs  = "req-args"
	kRequired = "required"
	kType     = "type"
)

/* =========================
   YAML structs
   ========================= */

type ServiceYAML struct {
	Required []string       `yaml:"required"`
	Type     string         `yaml:"type"`
	Rpm      int            `yaml:"rpm"`
	ReqArgs  map[string]any `yaml:"req-args"`
}

type ServicesYAML map[string]*ServiceYAML

var servicesYAML = loadServicesYAML()

func loadServicesYAML() ServicesYAML {
	data, err := embedFS.ReadFile("services.yaml")
	if err != nil {
		log.Fatalf("error reading services.yaml: %v", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Fatalf("error unmarshalling services.yaml: %v", err)
	}

	svcs := make(ServicesYAML, len(raw))
	for name, v := range raw {
		switch val := v.(type) {
		case nil, string:
			svcs[name] = nil
		case map[string]any:
			svcs[name] = parseServiceYAML(val)
		default:
			log.Fatalf("unexpected config for service %s: %T", name, v)
		}
	}

	return svcs
}

func parseServiceYAML(m map[string]any) *ServiceYAML {
	svc := &ServiceYAML{}
	if arrOfAny, ok := m[kRequired].([]any); ok {
		svc.Required = make([]string, 0, len(arrOfAny))
		for _, a := range arrOfAny {
			if s, ok := a.(string); ok {
				svc.Required = append(svc.Required, s)
			}
		}
	}
	if v, ok := m[kType].(string); ok {
		svc.Type = v
	}
	if v, ok := m[kRpm].(int); ok {
		svc.Rpm = v
	}
	if v, ok := m[kReqArgs].(map[string]any); ok {
		svc.ReqArgs = v
	}
	return svc
}

/* =========================
   Runtime ServiceConfig
   ========================= */

type ServiceConfig struct {
	Name  string
	Model string
	Type  string
	YAML  *ServiceYAML
}

func NewServiceConfig(service string) *ServiceConfig {
	svc := &ServiceConfig{Name: service}

	svcYAML, ok := servicesYAML[service]
	if !ok {
		// 若不存在服务的配置项，就默认为OpenAI兼容格式
		svc.setToOpenAICompatible()
		return svc
	}

	if svcYAML == nil { // 比如google服务
		return svc
	}

	// OpenAI兼容类型：先拷贝openai配置
	if svcYAML.Type == kOpenAI {
		svc.setToOpenAICompatible()
	}

	// 自身的配置
	svc.YAML = mergeConfig(svc.YAML, svcYAML)
	return svc
}

/* =========================
   Config helpers
   ========================= */

func copyConfig(svc *ServiceYAML) *ServiceYAML {
	if svc == nil {
		return nil
	}

	c := &ServiceYAML{
		Type:     svc.Type,
		Rpm:      svc.Rpm,
		Required: append([]string(nil), svc.Required...),
	}

	if svc.ReqArgs != nil {
		c.ReqArgs = make(map[string]any, len(svc.ReqArgs))
		maps.Copy(c.ReqArgs, svc.ReqArgs)
	}

	return c
}

func mergeConfig(base, override *ServiceYAML) *ServiceYAML {
	if base == nil {
		return copyConfig(override)
	}
	if override == nil {
		return copyConfig(base)
	}

	merged := copyConfig(base)

	// 使用反射通用合并非零字段
	if override.Type != "" {
		merged.Type = override.Type
	}
	if override.Rpm > 0 {
		merged.Rpm = override.Rpm
	}
	if len(override.Required) > 0 {
		merged.Required = append([]string(nil), override.Required...)
	}
	if override.ReqArgs != nil {
		if merged.ReqArgs == nil {
			merged.ReqArgs = map[string]any{}
		}
		maps.Copy(merged.ReqArgs, override.ReqArgs)
	}

	return merged
}

func (svc *ServiceConfig) setToOpenAICompatible() {
	svc.Type = kOpenAI
	svc.YAML = copyConfig(servicesYAML[kOpenAI])
}

/* =========================
   Env helpers
   ========================= */

func (svc *ServiceConfig) envKey(key string) string {
	return strings.ToUpper(
		strings.ReplaceAll(svc.Name+"_"+key, "-", "_"),
	)
}

func (svc *ServiceConfig) getEnvValue(key string) string {
	return os.Getenv(svc.envKey(key))
}

func (svc *ServiceConfig) GetEnvValue(key string) string {
	return svc.getEnvValue(key)
}

/* =========================
   Validation
   ========================= */

func (svc *ServiceConfig) ValidateEnvArgs() error {
	if svc.YAML == nil || len(svc.YAML.Required) == 0 {
		return nil
	}

	for _, key := range svc.YAML.Required {
		if svc.getEnvValue(key) == "" {
			return fmt.Errorf("%s", svc.getEnvArgsInfo())
		}
	}
	return nil
}

func (svc *ServiceConfig) getEnvArgsInfo() string {
	var lines []string

	lines = append(lines, "# Option 1: Set in a .env file (recommended)")
	if svc.YAML != nil {
		for _, key := range svc.YAML.Required {
			envKey := svc.envKey(key)
			lines = append(lines, fmt.Sprintf("%s=%q", envKey, os.Getenv(envKey)))
		}
	}

	lines = append(lines, "", "# Option 2: Export directly in your shell")
	if svc.YAML != nil {
		for _, key := range svc.YAML.Required {
			envKey := svc.envKey(key)
			lines = append(lines, fmt.Sprintf("export %s=%q", envKey, os.Getenv(envKey)))
		}
	}

	return strings.Join(lines, "\n")
}

/* =========================
   Runtime getters
   ========================= */

func (svc *ServiceConfig) GetReqArgs() map[string]any {
	if s := svc.getEnvValue(kReqArgs); s != "" {
		var v map[string]any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			return v
		} else {
			log.Printf("Warning: failed to parse req-args for %s: %v", svc.Name, err)
		}
	}

	if svc.YAML != nil {
		return svc.YAML.ReqArgs
	}
	return nil
}

func (svc *ServiceConfig) GetRpm() int {
	if s := svc.getEnvValue(kRpm); s != "" {
		if rpm, err := strconv.Atoi(s); err == nil {
			return rpm
		} else {
			log.Printf("Warning: failed to parse rpm for %s: %v", svc.Name, err)
		}
	}

	if svc.YAML != nil && svc.YAML.Rpm > 0 {
		return svc.YAML.Rpm
	}
	return 60
}

/* =========================
   Utilities
   ========================= */

func GetAllServiceNames() []string {
	names := make([]string, 0, len(servicesYAML))
	for name := range servicesYAML {
		names = append(names, name)
	}
	return names
}
