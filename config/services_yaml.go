package config

import (
	"log"
	"maps"

	"gopkg.in/yaml.v3"
)

const (
	kRequired       = "required"
	kType           = "type"
	kOpenAI         = "openai"
	kRequestArgs    = "request-args"
	kRpm            = "rpm"
	kMaxConcurrency = "max-concurrency"
)

/* =========================
   YAML structs
   ========================= */

type ServiceYAML struct {
	Required       []string       `yaml:"required"`
	Type           string         `yaml:"type"`
	Rpm            int            `yaml:"rpm"`
	MaxConcurrency int            `yaml:"max-concurrency"`
	RequestArgs    map[string]any `yaml:"request-args"`
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
	if v, ok := m[kMaxConcurrency].(int); ok {
		svc.MaxConcurrency = v
	}
	if v, ok := m[kRequestArgs].(map[string]any); ok {
		svc.RequestArgs = v
	}
	return svc
}

func (svc *ServiceYAML) copy() *ServiceYAML {
	if svc == nil {
		return nil
	}

	c := &ServiceYAML{
		Type:           svc.Type,
		Rpm:            svc.Rpm,
		MaxConcurrency: svc.MaxConcurrency,
		Required:       append([]string(nil), svc.Required...),
	}

	if svc.RequestArgs != nil {
		c.RequestArgs = make(map[string]any, len(svc.RequestArgs))
		maps.Copy(c.RequestArgs, svc.RequestArgs)
	}

	return c
}

func merge(base, override *ServiceYAML) *ServiceYAML {
	if base == nil {
		return override.copy()
	}
	if override == nil {
		return base.copy()
	}

	merged := base.copy()
	if override.Type != "" {
		merged.Type = override.Type
	}
	if override.Rpm > 0 {
		merged.Rpm = override.Rpm
	}
	if override.MaxConcurrency >= 0 {
		merged.MaxConcurrency = override.MaxConcurrency
	}
	if len(override.Required) > 0 {
		merged.Required = append([]string(nil), override.Required...)
	}
	if override.RequestArgs != nil {
		merged.RequestArgs = maps.Clone(override.RequestArgs)
	}
	return merged
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
