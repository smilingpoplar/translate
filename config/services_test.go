package config

import (
	"os"
	"testing"
)

func TestGetRpm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		service  string
		envRpm   string
		wantRpm  int
		clearEnv bool
	}{
		{
			name:    "openai default rpm",
			service: "openai",
			wantRpm: 60,
		},
		{
			name:    "siliconflow default rpm",
			service: "siliconflow",
			wantRpm: 300,
		},
		{
			name:    "glm inherits openai rpm",
			service: "glm",
			wantRpm: 60,
		},
		{
			name:    "openai env override",
			service: "openai",
			envRpm:  "120",
			wantRpm: 120,
		},
		{
			name:    "siliconflow env override",
			service: "siliconflow",
			envRpm:  "500",
			wantRpm: 500,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewServiceConfig(tt.service)

			// 清理环境变量
			if tt.clearEnv || tt.envRpm != "" {
				envKey := svc.envKey("rpm")
				original := os.Getenv(envKey)
				t.Cleanup(func() {
					if original == "" {
						os.Unsetenv(envKey)
					} else {
						os.Setenv(envKey, original)
					}
				})

				if tt.envRpm != "" {
					os.Setenv(envKey, tt.envRpm)
				}
			}
			got := svc.GetRpm()

			if got != tt.wantRpm {
				t.Errorf("GetRpm() = %d, want %d", got, tt.wantRpm)
			}
		})
	}
}

func TestGetReqArgs(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		envArgs  string
		wantKeys []string
	}{
		{
			name:     "openai no req-args",
			service:  "openai",
			wantKeys: []string{},
		},
		{
			name:     "glm has req-args",
			service:  "glm",
			wantKeys: []string{"thinking"},
		},
		{
			name:     "siliconflow has req-args",
			service:  "siliconflow",
			wantKeys: []string{"enable_thinking"},
		},
		{
			name:     "env override req-args",
			service:  "openai",
			envArgs:  `{"temperature": 0.5, "top_p": 0.9}`,
			wantKeys: []string{"temperature", "top_p"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewServiceConfig(tt.service)

			if tt.envArgs != "" {
				envKey := svc.envKey("req-args")
				original := os.Getenv(envKey)
				t.Cleanup(func() {
					if original == "" {
						os.Unsetenv(envKey)
					} else {
						os.Setenv(envKey, original)
					}
				})
				os.Setenv(envKey, tt.envArgs)
			}
			got := svc.GetReqArgs()

			if len(got) != len(tt.wantKeys) {
				t.Errorf("GetReqArgs() keys = %v, want %v", keysOf(got), tt.wantKeys)
			}

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("GetReqArgs() missing key: %s", key)
				}
			}
		})
	}
}

func TestValidateEnvArgs(t *testing.T) {
	tests := []struct {
		name       string
		service    string
		setEnvVars map[string]string
		wantErr    bool
	}{
		{
			name:    "openai missing required env",
			service: "openai",
			wantErr: true,
		},
		{
			name:       "openai with all required env",
			service:    "openai",
			setEnvVars: map[string]string{"OPENAI_BASE_URL": "http://example.com", "OPENAI_API_KEY": "key", "OPENAI_MODEL": "gpt-4"},
			wantErr:    false,
		},
		{
			name:    "google no required env",
			service: "google",
			wantErr: false,
		},
		{
			name:    "glm inherits openai required",
			service: "glm",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cleanupKeys []string
			for k, v := range tt.setEnvVars {
				cleanupKeys = append(cleanupKeys, k)
				original := os.Getenv(k)
				t.Cleanup(func() {
					if original == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, original)
					}
				})
				os.Setenv(k, v)
			}

			svc := NewServiceConfig(tt.service)
			err := svc.ValidateEnvArgs()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
