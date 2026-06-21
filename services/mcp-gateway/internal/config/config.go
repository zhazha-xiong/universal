package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var placeholderPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Config describes the gateway runtime configuration.
type Config struct {
	Server Server `yaml:"server"`
	MySQL  MySQL  `yaml:"mysql"`
}

// Server describes HTTP server configuration.
type Server struct {
	Addr string `yaml:"addr"`
}

// MySQL describes MySQL connection configuration.
type MySQL struct {
	DSN string `yaml:"dsn"`
}

// Load reads a YAML config file and expands ${ENV_NAME} placeholders.
func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	expanded, err := expandEnv(string(content))
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func expandEnv(content string) (string, error) {
	var missing string
	expanded := placeholderPattern.ReplaceAllStringFunc(content, func(match string) string {
		name := placeholderPattern.FindStringSubmatch(match)[1]
		value, ok := os.LookupEnv(name)
		if !ok {
			missing = name
			return match
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("environment variable %s is not set", missing)
	}
	return expanded, nil
}
