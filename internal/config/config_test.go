package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test ServerConfig defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host: expected 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port: expected 8080, got %d", cfg.Server.Port)
	}

	// Test BINDConfig defaults
	if cfg.BIND.NamedConfPath != "/etc/bind/named.conf" {
		t.Errorf("BIND.NamedConfPath: expected /etc/bind/named.conf, got %s", cfg.BIND.NamedConfPath)
	}
	if cfg.BIND.ZoneDirectory != "./zones" {
		t.Errorf("BIND.ZoneDirectory: expected ./zones, got %s", cfg.BIND.ZoneDirectory)
	}
	if cfg.BIND.RndcPath != "/usr/sbin/rndc" {
		t.Errorf("BIND.RndcPath: expected /usr/sbin/rndc, got %s", cfg.BIND.RndcPath)
	}
	if cfg.BIND.RndcConfPath != "/etc/bind/rndc.conf" {
		t.Errorf("BIND.RndcConfPath: expected /etc/bind/rndc.conf, got %s", cfg.BIND.RndcConfPath)
	}
	if cfg.BIND.DefaultTTL != 3600 {
		t.Errorf("BIND.DefaultTTL: expected 3600, got %d", cfg.BIND.DefaultTTL)
	}
	if cfg.BIND.DefaultRefresh != 7200 {
		t.Errorf("BIND.DefaultRefresh: expected 7200, got %d", cfg.BIND.DefaultRefresh)
	}
	if cfg.BIND.DefaultRetry != 3600 {
		t.Errorf("BIND.DefaultRetry: expected 3600, got %d", cfg.BIND.DefaultRetry)
	}
	if cfg.BIND.DefaultExpire != 1209600 {
		t.Errorf("BIND.DefaultExpire: expected 1209600, got %d", cfg.BIND.DefaultExpire)
	}
	if cfg.BIND.DefaultMinimum != 86400 {
		t.Errorf("BIND.DefaultMinimum: expected 86400, got %d", cfg.BIND.DefaultMinimum)
	}

	// Test LoggingConfig defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level: expected info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format: expected json, got %s", cfg.Logging.Format)
	}
	if cfg.Logging.OutputPath != "stdout" {
		t.Errorf("Logging.OutputPath: expected stdout, got %s", cfg.Logging.OutputPath)
	}
}

func TestLoadConfigNonExistentFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("LoadConfig should return nil error for non-existent file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig should return default config for non-existent file")
	}

	// Verify it returns defaults
	if cfg.Server.Port != 8080 {
		t.Error("Should return default config")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("LoadConfig should return error for invalid JSON")
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid.json")

	configJSON := `{
		"server": {
			"host": "127.0.0.1",
			"port": 9090
		},
		"bind": {
			"named_conf_path": "/custom/named.conf",
			"zone_directory": "/custom/zones",
			"rndc_path": "/custom/rndc",
			"default_ttl": 1800
		},
		"logging": {
			"level": "debug",
			"format": "text",
			"output_path": "/var/log/app.log"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify custom values
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host: expected 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port: expected 9090, got %d", cfg.Server.Port)
	}
	if cfg.BIND.NamedConfPath != "/custom/named.conf" {
		t.Errorf("BIND.NamedConfPath: expected /custom/named.conf, got %s", cfg.BIND.NamedConfPath)
	}
	if cfg.BIND.ZoneDirectory != "/custom/zones" {
		t.Errorf("BIND.ZoneDirectory: expected /custom/zones, got %s", cfg.BIND.ZoneDirectory)
	}
	if cfg.BIND.DefaultTTL != 1800 {
		t.Errorf("BIND.DefaultTTL: expected 1800, got %d", cfg.BIND.DefaultTTL)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level: expected debug, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format: expected text, got %s", cfg.Logging.Format)
	}
}

func TestLoadConfigPartialFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.json")

	// Only override port, rest should be defaults
	configJSON := `{
		"server": {
			"port": 3000
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify overridden value
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port: expected 3000, got %d", cfg.Server.Port)
	}

	// Verify defaults are preserved
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host: expected 0.0.0.0 (default), got %s", cfg.Server.Host)
	}
	if cfg.BIND.DefaultTTL != 3600 {
		t.Errorf("BIND.DefaultTTL: expected 3600 (default), got %d", cfg.BIND.DefaultTTL)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved.json")

	cfg := &Config{
		Server: ServerConfig{
			Host: "192.168.1.1",
			Port: 8888,
		},
		BIND: BINDConfig{
			ZoneDirectory:  "/data/zones",
			DefaultTTL:     7200,
			DefaultRefresh: 14400,
		},
		Logging: LoggingConfig{
			Level:  "warn",
			Format: "json",
		},
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal saved config: %v", err)
	}

	if loaded.Server.Host != cfg.Server.Host {
		t.Errorf("Server.Host: expected %s, got %s", cfg.Server.Host, loaded.Server.Host)
	}
	if loaded.Server.Port != cfg.Server.Port {
		t.Errorf("Server.Port: expected %d, got %d", cfg.Server.Port, loaded.Server.Port)
	}
	if loaded.BIND.ZoneDirectory != cfg.BIND.ZoneDirectory {
		t.Errorf("BIND.ZoneDirectory: expected %s, got %s", cfg.BIND.ZoneDirectory, loaded.BIND.ZoneDirectory)
	}
	if loaded.Logging.Level != cfg.Logging.Level {
		t.Errorf("Logging.Level: expected %s, got %s", cfg.Logging.Level, loaded.Logging.Level)
	}
}

func TestSaveConfigAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip.json")

	original := DefaultConfig()
	original.Server.Port = 4444
	original.BIND.DefaultTTL = 9999

	// Save
	if err := SaveConfig(original, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Compare
	if loaded.Server.Port != original.Server.Port {
		t.Errorf("Server.Port: expected %d, got %d", original.Server.Port, loaded.Server.Port)
	}
	if loaded.BIND.DefaultTTL != original.BIND.DefaultTTL {
		t.Errorf("BIND.DefaultTTL: expected %d, got %d", original.BIND.DefaultTTL, loaded.BIND.DefaultTTL)
	}
}

func TestConfigJSONMarshaling(t *testing.T) {
	cfg := DefaultConfig()

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal Config: %v", err)
	}

	var unmarshaled Config
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Config: %v", err)
	}

	if unmarshaled.Server.Host != cfg.Server.Host {
		t.Errorf("Server.Host mismatch")
	}
	if unmarshaled.Server.Port != cfg.Server.Port {
		t.Errorf("Server.Port mismatch")
	}
	if unmarshaled.BIND.ZoneDirectory != cfg.BIND.ZoneDirectory {
		t.Errorf("BIND.ZoneDirectory mismatch")
	}
}

func TestServerConfigMarshaling(t *testing.T) {
	cfg := ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal ServerConfig: %v", err)
	}

	var unmarshaled ServerConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ServerConfig: %v", err)
	}

	if unmarshaled.Host != cfg.Host {
		t.Errorf("Host mismatch: expected %s, got %s", cfg.Host, unmarshaled.Host)
	}
	if unmarshaled.Port != cfg.Port {
		t.Errorf("Port mismatch: expected %d, got %d", cfg.Port, unmarshaled.Port)
	}
}

func TestBINDConfigMarshaling(t *testing.T) {
	cfg := BINDConfig{
		NamedConfPath:  "/etc/bind/named.conf",
		ZoneDirectory:  "./zones",
		RndcPath:       "/usr/sbin/rndc",
		RndcConfPath:   "/etc/bind/rndc.conf",
		DefaultTTL:     3600,
		DefaultRefresh: 7200,
		DefaultRetry:   3600,
		DefaultExpire:  1209600,
		DefaultMinimum: 86400,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal BINDConfig: %v", err)
	}

	var unmarshaled BINDConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal BINDConfig: %v", err)
	}

	if unmarshaled.ZoneDirectory != cfg.ZoneDirectory {
		t.Errorf("ZoneDirectory mismatch")
	}
	if unmarshaled.DefaultTTL != cfg.DefaultTTL {
		t.Errorf("DefaultTTL mismatch")
	}
}

func TestLoggingConfigMarshaling(t *testing.T) {
	cfg := LoggingConfig{
		Level:      "debug",
		Format:     "text",
		OutputPath: "/var/log/app.log",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal LoggingConfig: %v", err)
	}

	var unmarshaled LoggingConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal LoggingConfig: %v", err)
	}

	if unmarshaled.Level != cfg.Level {
		t.Errorf("Level mismatch: expected %s, got %s", cfg.Level, unmarshaled.Level)
	}
	if unmarshaled.Format != cfg.Format {
		t.Errorf("Format mismatch: expected %s, got %s", cfg.Format, unmarshaled.Format)
	}
}
