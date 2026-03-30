package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server         ServerConfig         `json:"server"`
	BIND           BINDConfig           `json:"bind"`
	Logging        LoggingConfig        `json:"logging"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// BINDConfig represents BIND DNS configuration
type BINDConfig struct {
	NamedConfPath    string `json:"named_conf_path"`
	ZoneDirectory    string `json:"zone_directory"`
	RndcPath         string `json:"rndc_path"`
	RndcConfPath     string `json:"rndc_conf_path"`
	DefaultTTL       int    `json:"default_ttl"`
	DefaultRefresh   int    `json:"default_refresh"`
	DefaultRetry     int    `json:"default_retry"`
	DefaultExpire    int    `json:"default_expire"`
	DefaultMinimum   int    `json:"default_minimum"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"output_path"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		BIND: BINDConfig{
			NamedConfPath:  "/etc/bind/named.conf",
			ZoneDirectory:  "./zones",
			RndcPath:       "/usr/sbin/rndc",
			RndcConfPath:   "/etc/bind/rndc.conf",
			DefaultTTL:     3600,
			DefaultRefresh: 7200,
			DefaultRetry:   3600,
			DefaultExpire:  1209600,
			DefaultMinimum: 86400,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
