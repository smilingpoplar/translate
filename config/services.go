package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

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
	svc.YAML = merge(svc.YAML, svcYAML)
	return svc
}

func (svc *ServiceConfig) setToOpenAICompatible() {
	svc.Type = kOpenAI
	svc.YAML = servicesYAML[kOpenAI].copy()
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
// 环境变量配置JSON串 => map
func (svc *ServiceConfig) GetRequestArgs() map[string]any {
	if s := svc.getEnvValue(kRequestArgs); s != "" {
		var v map[string]any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			return v
		} else {
			log.Printf("Warning: failed to parse request-args for %s: %v", svc.Name, err)
		}
	}

	if svc.YAML != nil {
		return svc.YAML.RequestArgs
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

func (svc *ServiceConfig) GetMaxConcurrency() int {
	if s := svc.getEnvValue(kMaxConcurrency); s != "" {
		if conc, err := strconv.Atoi(s); err == nil {
			return conc
		} else {
			log.Printf("Warning: failed to parse max-concurrency for %s: %v", svc.Name, err)
		}
	}

	if svc.YAML != nil && svc.YAML.MaxConcurrency > 0 {
		return svc.YAML.MaxConcurrency
	}
	return 0 // 0 表示不限制
}
