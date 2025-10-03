package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	kType     = "type"
	kOpenAI   = "openai"
)

type ServiceConfig struct {
	Name      string
	Model     string
	ConfigMap map[string]any
	Type      string
}

func NewServiceConfig(service string) *ServiceConfig {
	sc := &ServiceConfig{}
	// name
	sc.Name = service
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
	} else {
		sc.ConfigMap = configMap
	}
	return sc
}

func (sc *ServiceConfig) setToOpenAICompatible() {
	sc.Type = kOpenAI
	configOpenAI := servicesConfig[kOpenAI].(map[string]any)
	sc.ConfigMap = configOpenAI
}

func (sc *ServiceConfig) ValidateEnvArgs() error {
	for _, k := range sc.ConfigMap[kRequired].([]any) {
		if sc.GetEnvValue(k.(string)) == "" {
			return fmt.Errorf("%s", sc.getEnvArgsInfo())
		}
	}

	return nil
}

func (sc *ServiceConfig) getEnvArgsInfo() string {
	msg := "# Option 1: Set in a .env file (recommended)"
	for _, k := range sc.ConfigMap[kRequired].([]any) {
		key := sc.getEnvKey(k.(string))
		val := os.Getenv(key)
		msg += fmt.Sprintf("\n%s=%q", key, val)
	}
	msg += "\n\n# Option 2: Export directly in your shell"
	for _, k := range sc.ConfigMap[kRequired].([]any) {
		key := sc.getEnvKey(k.(string))
		val := os.Getenv(key)
		msg += fmt.Sprintf("\nexport %s=%q", key, val)
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

func GetAllServiceNames() []string {
	var names []string
	for service := range servicesConfig {
		names = append(names, service)
	}
	return names
}
