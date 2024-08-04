package config

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed services.yaml
var embedFS embed.FS

var servicesConfig = getServicesConfig()

func getServicesConfig() map[string]any {
	data, err := embedFS.ReadFile("services.yaml")
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	var config map[string]any
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("error unmarshalling YAML: %v", err)
	}
	return config
}

const (
	kRequired = "required"
	kOptional = "optional"
	kDefault  = "default"
	kType     = "type"
	kOpenAI   = "openai"
	kModel    = "model"
)

type ServiceConfig struct {
	Name      string
	Model     string
	ConfigMap map[string]any
	Type      string
}

func NewServiceConfig(service string) *ServiceConfig {
	sc := &ServiceConfig{}
	// name, model
	parts := strings.SplitN(service, ":", 2)
	sc.Name = parts[0]
	if len(parts) == 2 {
		sc.Model = parts[1]
	}

	serviceConfig, ok := servicesConfig[sc.Name]
	if !ok { // 若不存在服务的配置项，就默认为OpenAI兼容格式
		sc.setToOpenAICompatible()
		return sc
	}

	// service config map
	configMap, ok := serviceConfig.(map[string]any)
	if !ok { // 比如google服务，不需要configMap
		return sc
	}

	if configMap[kType] == kOpenAI { // OpenAI兼容类型
		sc.setToOpenAICompatible()
		sc.ConfigMap[kDefault] = configMap[kDefault]
	} else {
		sc.ConfigMap = configMap
	}
	return sc
}

func (sc *ServiceConfig) setToOpenAICompatible() {
	sc.Type = kOpenAI
	configOpenAI := servicesConfig[kOpenAI].(map[string]any)
	configOpenAI[kDefault] = make(map[string]any)
	sc.ConfigMap = configOpenAI
}

func (sc *ServiceConfig) ValidateEnvArgs() error {
	if sc.Model != "" {
		os.Setenv(sc.getEnvKey(kModel), sc.Model)
	}
	defaultSettings := sc.ConfigMap[kDefault].(map[string]any)
	for k, v := range defaultSettings {
		if v != nil {
			key := sc.getEnvKey(k)
			if val := os.Getenv(key); val == "" {
				os.Setenv(key, v.(string))
			}
		}
	}

	for _, k := range sc.ConfigMap[kRequired].([]any) {
		if sc.GetEnvValue(k.(string)) == "" {
			return fmt.Errorf("%s", sc.getEnvArgsInfo())
		}
	}

	return nil
}

func (sc *ServiceConfig) getEnvArgsInfo() string {
	msg := "Please set the environment variables required for this service."
	for _, sec := range []string{kRequired, kOptional} {
		msg += fmt.Sprintf("\n### %s", sec)
		for _, k := range sc.ConfigMap[sec].([]any) {
			key := sc.getEnvKey(k.(string))
			val := os.Getenv(key)
			msg += fmt.Sprintf("\n%s=%q", key, val)
		}
	}
	return msg
}

func (sc *ServiceConfig) getEnvKey(key string) string {
	str := fmt.Sprintf("%s_%s", sc.Name, key)
	str = strings.ReplaceAll(str, "-", "_")
	str = strings.ToUpper(str)
	return str
}

func (sc *ServiceConfig) GetEnvValue(key string) string {
	k := sc.getEnvKey(key)
	return os.Getenv(k)
}

func GetAllServiceStrs() []string {
	var names []string
	for service := range servicesConfig {
		defaultModel := getDefaultModel(service)
		if defaultModel != "" {
			service = fmt.Sprintf("%s:%s", service, defaultModel)
		}
		names = append(names, service)
	}
	return names
}

func getDefaultModel(service string) string {
	configMap, ok := servicesConfig[service].(map[string]any)
	if !ok {
		return ""
	}
	defaultSettings, ok := configMap[kDefault].(map[string]any)
	if !ok {
		return ""
	}
	return defaultSettings[kModel].(string)
}
