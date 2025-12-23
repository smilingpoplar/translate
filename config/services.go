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
		sc.ConfigMap = make(map[string]any)
		sc.setToOpenAICompatible()
		return sc
	}

	// service config map
	configMap, ok := serviceConfig.(map[string]any)
	if !ok { // 比如google服务，不需要configMap
		return sc
	}

	sc.ConfigMap = make(map[string]any)
	if configMap[kType] == kOpenAI { // OpenAI兼容类型
		sc.setToOpenAICompatible()
	}
	maps.Copy(sc.ConfigMap, configMap) // 自身的配置
	return sc
}

func (sc *ServiceConfig) setToOpenAICompatible() {
	sc.Type = kOpenAI
	configOpenAI := servicesConfig[kOpenAI].(map[string]any)
	maps.Copy(sc.ConfigMap, configOpenAI)
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

func (sc *ServiceConfig) GetReqArgs() map[string]any {
	// 优先检查环境变量，环境变量以json串配置，如 XXX_REQ_ARGS='{"enable_thinking": false}'
	if reqArgsStr := sc.GetEnvValue("req-args"); reqArgsStr != "" {
		var reqArgs map[string]any
		if err := json.Unmarshal([]byte(reqArgsStr), &reqArgs); err == nil {
			return reqArgs
		} else {
			log.Printf("Warning: failed to parse req-args from environment variable for %s: %v", sc.Name, err)
		}
	}

	// 如果环境变量不存在或解析失败，使用配置文件中的配置
	if req, ok := sc.ConfigMap["req-args"].(map[string]any); ok {
		return req
	}
	return nil
}

func (sc *ServiceConfig) GetRpm() int {
	if rpmStr := sc.GetEnvValue("rpm"); rpmStr != "" {
		if rpm, err := strconv.Atoi(rpmStr); err == nil {
			return rpm
		} else {
			log.Printf("Warning: failed to parse rpm from environment variable for %s: %v", sc.Name, err)
		}
	}

	// 如果环境变量不存在或解析失败，使用配置文件中的配置
	if rpm, ok := sc.ConfigMap["rpm"].(int); ok {
		return rpm
	}
	return 60
}

func GetAllServiceNames() []string {
	var names []string
	for service := range servicesConfig {
		names = append(names, service)
	}
	return names
}
