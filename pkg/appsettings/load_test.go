package appsettings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type TestConfig struct {
	DatabaseURL string  `json:"databaseURL"`
	Port        int     `json:"port"`
	DebugMode   bool    `json:"debugMode"`
	Timeout     float64 `json:"timeout"`
	Name        string  `json:"name"`
}

func TestNew(t *testing.T) {
	appSettings := New[TestConfig]()

	if appSettings == nil {
		t.Fatal("New() returned nil")
	}

	if appSettings.withArgs != nil {
		t.Error("Expected withArgs to be nil")
	}

	if appSettings.withEnvVars != nil {
		t.Error("Expected withEnvVars to be nil")
	}

	if appSettings.withEnvironment != nil {
		t.Error("Expected withEnvironment to be nil")
	}

	if appSettings.withConfigDirectory != nil {
		t.Error("Expected withConfigDirectory to be nil")
	}
}

func TestWithArgs(t *testing.T) {
	appSettings := New[TestConfig]()
	args := []string{"--port", "8080", "--debug"}

	result := appSettings.WithArgs(args)

	if result != appSettings {
		t.Error("WithArgs should return the same instance for chaining")
	}

	if !reflect.DeepEqual(appSettings.withArgs, args) {
		t.Errorf("Expected args %v, got %v", args, appSettings.withArgs)
	}
}

func TestWithEnvVars(t *testing.T) {
	appSettings := New[TestConfig]()
	envVars := []string{"PORT=8080", "DEBUGMODE=true"}

	result := appSettings.WithEnvVars(envVars)

	if result != appSettings {
		t.Error("WithEnvVars should return the same instance for chaining")
	}

	if !reflect.DeepEqual(appSettings.withEnvVars, envVars) {
		t.Errorf("Expected env vars %v, got %v", envVars, appSettings.withEnvVars)
	}
}

func TestWithEnvironment(t *testing.T) {
	appSettings := New[TestConfig]()
	environment := "dev"

	result := appSettings.WithEnvironment(environment)

	if result != appSettings {
		t.Error("WithEnvironment should return the same instance for chaining")
	}

	if appSettings.withEnvironment == nil || *appSettings.withEnvironment != environment {
		t.Errorf("Expected environment %s, got %v", environment, appSettings.withEnvironment)
	}
}

func TestWithConfigDirectory(t *testing.T) {
	appSettings := New[TestConfig]()
	configDir := "/tmp/config"

	result := appSettings.WithConfigDirectory(configDir)

	if result != appSettings {
		t.Error("WithConfigDirectory should return the same instance for chaining")
	}

	if appSettings.withConfigDirectory == nil || *appSettings.withConfigDirectory != configDir {
		t.Errorf("Expected config directory %s, got %v", configDir, appSettings.withConfigDirectory)
	}
}

func TestGetWD(t *testing.T) {
	appSettings := New[TestConfig]()

	wd, err := appSettings.getWD()
	if err != nil {
		t.Fatalf("getWD() returned error: %v", err)
	}

	if wd == nil || *wd == "" {
		t.Error("getWD() should return a non-empty directory path")
	}
}

func TestGetConfigDirectory_WithCustomDirectory(t *testing.T) {
	appSettings := New[TestConfig]()
	customDir := "/tmp/custom"
	appSettings.WithConfigDirectory(customDir)

	configDir, err := appSettings.getConfigDirectory()
	if err != nil {
		t.Fatalf("getConfigDirectory() returned error: %v", err)
	}

	if configDir == nil || *configDir != customDir {
		t.Errorf("Expected config directory %s, got %v", customDir, configDir)
	}
}

func TestGetConfigDirectory_FallbackToWD(t *testing.T) {
	appSettings := New[TestConfig]()

	configDir, err := appSettings.getConfigDirectory()
	if err != nil {
		t.Fatalf("getConfigDirectory() returned error: %v", err)
	}

	wd, err := appSettings.getWD()
	if err != nil {
		t.Fatalf("getWD() returned error: %v", err)
	}

	if configDir == nil || *configDir != *wd {
		t.Errorf("Expected config directory to fall back to working directory %s, got %v", *wd, configDir)
	}
}

func TestLoadConfigFile_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	testConfig := map[string]interface{}{
		"databaseURL": "postgres://localhost/test",
		"port":        float64(8080),
		"debugMode":   true,
	}

	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err = appSettings.loadConfigFile(configFile, configMap)
	if err != nil {
		t.Fatalf("loadConfigFile() returned error: %v", err)
	}

	expectedKeys := []string{"databaseURL", "port", "debugMode"}
	for _, key := range expectedKeys {
		if configMap[key] != testConfig[key] {
			t.Errorf("For key %s: expected %v, got %v", key, testConfig[key], configMap[key])
		}
	}
}

func TestLoadConfigFile_FileNotExists(t *testing.T) {
	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err := appSettings.loadConfigFile("/nonexistent/config.json", configMap)
	if err != nil {
		t.Errorf("loadConfigFile() should not return error for non-existent file, got: %v", err)
	}

	if len(configMap) != 0 {
		t.Errorf("configMap should be empty when file doesn't exist, got: %v", configMap)
	}
}

func TestLoadConfigFile_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	if err := os.WriteFile(configFile, []byte("{invalid json}"), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err := appSettings.loadConfigFile(configFile, configMap)
	if err == nil {
		t.Error("loadConfigFile() should return error for invalid JSON")
	}
}

func TestLoadEnvVars(t *testing.T) {
	appSettings := New[TestConfig]()
	envVars := []string{
		"PORT=8080",
		"DATABASEURL=postgres://localhost/test",
		"DEBUGMODE=true",
		"TIMEOUT=30.5",
		"NAME=test-app",
		"INVALID_VAR",
		"ANOTHER=INVALID",
	}
	appSettings.WithEnvVars(envVars)

	configMap := make(map[string]interface{})
	err := appSettings.loadEnvVars(configMap)
	if err != nil {
		t.Fatalf("loadEnvVars() returned error: %v", err)
	}

	expected := map[string]interface{}{
		"port":        8080,
		"databaseurl": "postgres://localhost/test",
		"debugmode":   true,
		"timeout":     30.5,
		"name":        "test-app",
		"another":     "INVALID",
	}

	if !reflect.DeepEqual(configMap, expected) {
		t.Errorf("Expected config %v, got %v", expected, configMap)
	}
}

func TestLoadEnvVars_NoEnvVars(t *testing.T) {
	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err := appSettings.loadEnvVars(configMap)
	if err != nil {
		t.Fatalf("loadEnvVars() returned error: %v", err)
	}

	if len(configMap) != 0 {
		t.Errorf("configMap should be empty when no env vars are set, got: %v", configMap)
	}
}

func TestLoadEnvVars_FullIntegration(t *testing.T) {
	tempDir := t.TempDir()

	appSettings := New[TestConfig]().
		WithConfigDirectory(tempDir).
		WithEnvVars([]string{
			"PORT=3000",
			"DATABASEURL=postgres://localhost/env-test",
			"DEBUGMODE=true",
			"TIMEOUT=25.5",
			"NAME=env-test-app",
		})

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := &TestConfig{
		DatabaseURL: "postgres://localhost/env-test",
		Port:        3000,
		DebugMode:   true,
		Timeout:     25.5,
		Name:        "env-test-app",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected config %+v, got %+v", expected, result)
	}
}

func TestLoadEnvVars_UnderscoreIgnored(t *testing.T) {
	tempDir := t.TempDir()

	appSettings := New[TestConfig]().
		WithConfigDirectory(tempDir).
		WithEnvVars([]string{
			"PORT=3000",
			"DATABASE_URL=postgres://localhost/should-ignore",
			"DEBUG_MODE=true",
			"TIMEOUT=25.5",
		})

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := &TestConfig{
		DatabaseURL: "",
		Port:        3000,
		DebugMode:   false,
		Timeout:     25.5,
		Name:        "",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected config %+v, got %+v", expected, result)
	}

	if result.DatabaseURL != "" {
		t.Errorf("DATABASE_URL should be ignored, but DatabaseURL = %q", result.DatabaseURL)
	}
	if result.DebugMode != false {
		t.Errorf("DEBUG_MODE should be ignored, but DebugMode = %v", result.DebugMode)
	}
}

func TestLoadArgs(t *testing.T) {
	appSettings := New[TestConfig]()
	args := []string{
		"program",
		"--port", "9000",
		"--debug-mode", "false",
		"--timeout", "45.5",
		"--name", "cli-app",
		"--verbose",
		"--another-flag",
		"regular-arg",
	}
	appSettings.WithArgs(args)

	configMap := make(map[string]interface{})
	err := appSettings.loadArgs(configMap)
	if err != nil {
		t.Fatalf("loadArgs() returned error: %v", err)
	}

	expected := map[string]interface{}{
		"port":         9000,
		"debug-mode":   false,
		"timeout":      45.5,
		"name":         "cli-app",
		"verbose":      true,
		"another-flag": "regular-arg",
	}

	if !reflect.DeepEqual(configMap, expected) {
		t.Errorf("Expected config %v, got %v", expected, configMap)
	}
}

func TestLoadArgs_NoArgs(t *testing.T) {
	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err := appSettings.loadArgs(configMap)
	if err != nil {
		t.Fatalf("loadArgs() returned error: %v", err)
	}

	if len(configMap) != 0 {
		t.Errorf("configMap should be empty when no args are set, got: %v", configMap)
	}
}

func TestParseValue(t *testing.T) {
	appSettings := New[TestConfig]()

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"true", true},
		{"false", false},
		{"True", true},
		{"FALSE", false},
		{"123", 123},
		{"-456", -456},
		{"0", false},
		{"123.45", 123.45},
		{"-67.89", -67.89},
		{"0.0", 0.0},
		{"hello", "hello"},
		{"", ""},
		{"123abc", "123abc"},
		{"true123", "true123"},
	}

	for _, test := range tests {
		result := appSettings.parseValue(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("parseValue(%q) = %v (%T), expected %v (%T)",
				test.input, result, result, test.expected, test.expected)
		}
	}
}

func TestUnmarshalToType_Success(t *testing.T) {
	appSettings := New[TestConfig]()
	configMap := map[string]interface{}{
		"databaseURL": "postgres://localhost/test",
		"port":        8080,
		"debugMode":   true,
		"timeout":     30.5,
		"name":        "test-app",
	}

	result, err := appSettings.unmarshalToType(configMap)
	if err != nil {
		t.Fatalf("unmarshalToType() returned error: %v", err)
	}

	expected := &TestConfig{
		DatabaseURL: "postgres://localhost/test",
		Port:        8080,
		DebugMode:   true,
		Timeout:     30.5,
		Name:        "test-app",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected config %+v, got %+v", expected, result)
	}
}

func TestUnmarshalToType_InvalidData(t *testing.T) {
	appSettings := New[TestConfig]()
	configMap := map[string]interface{}{
		"port": "invalid-port",
	}

	_, err := appSettings.unmarshalToType(configMap)
	if err == nil {
		t.Error("unmarshalToType() should return error for invalid data")
	}
}

func TestLoad_FullIntegration(t *testing.T) {
	tempDir := t.TempDir()

	baseConfig := map[string]interface{}{
		"databaseURL": "postgres://localhost/base",
		"port":        8000,
		"debugMode":   false,
		"timeout":     10.0,
		"name":        "base-app",
	}
	baseData, err := json.Marshal(baseConfig)
	if err != nil {
		t.Fatalf("Failed to marshal base config: %v", err)
	}
	baseConfigFile := filepath.Join(tempDir, "config.json")
	if err = os.WriteFile(baseConfigFile, baseData, 0600); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	envConfig := map[string]interface{}{
		"databaseURL": "postgres://localhost/dev",
		"debugMode":   true,
		"timeout":     20.0,
	}
	envData, err := json.Marshal(envConfig)
	if err != nil {
		t.Fatalf("Failed to marshal env config: %v", err)
	}
	envConfigFile := filepath.Join(tempDir, "config.dev.json")
	if err = os.WriteFile(envConfigFile, envData, 0600); err != nil {
		t.Fatalf("Failed to write env config: %v", err)
	}

	appSettings := New[TestConfig]().
		WithConfigDirectory(tempDir).
		WithEnvironment("dev").
		WithEnvVars([]string{
			"PORT=9000",
			"NAME=env-app",
		}).
		WithArgs([]string{
			"program",
			"--timeout", "45.5",
			"--debugmode", "false",
		})

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := &TestConfig{
		DatabaseURL: "postgres://localhost/dev",
		Port:        9000,
		DebugMode:   false,
		Timeout:     45.5,
		Name:        "env-app",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected config %+v, got %+v", expected, result)
	}
}

func TestLoad_NoConfigFiles(t *testing.T) {
	tempDir := t.TempDir()

	appSettings := New[TestConfig]().
		WithConfigDirectory(tempDir).
		WithEnvironment("nonexistent").
		WithEnvVars([]string{
			"PORT=3000",
			"DEBUGMODE=true",
		})

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := &TestConfig{
		DatabaseURL: "",
		Port:        3000,
		DebugMode:   true,
		Timeout:     0.0,
		Name:        "",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected config %+v, got %+v", expected, result)
	}
}

func TestLoad_InvalidConfigDirectory(t *testing.T) {
	appSettings := New[TestConfig]().
		WithConfigDirectory("/nonexistent/path/that/does/not/exist")

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := &TestConfig{}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected empty config %+v, got %+v", expected, result)
	}
}

func TestLoad_InvalidBaseConfig(t *testing.T) {
	tempDir := t.TempDir()

	baseConfigFile := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(baseConfigFile, []byte("{invalid}"), 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	appSettings := New[TestConfig]().WithConfigDirectory(tempDir)

	_, err := appSettings.Load()
	if err == nil {
		t.Error("Load() should return error for invalid base config")
	}
}

func TestLoad_InvalidEnvConfig(t *testing.T) {
	tempDir := t.TempDir()

	baseConfig := map[string]interface{}{"port": 8080}
	baseData, err := json.Marshal(baseConfig)
	if err != nil {
		t.Fatalf("Failed to marshal base config: %v", err)
	}
	baseConfigFile := filepath.Join(tempDir, "config.json")
	if err = os.WriteFile(baseConfigFile, baseData, 0600); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	envConfigFile := filepath.Join(tempDir, "config.dev.json")
	if err := os.WriteFile(envConfigFile, []byte("{invalid}"), 0600); err != nil {
		t.Fatalf("Failed to write invalid env config: %v", err)
	}

	appSettings := New[TestConfig]().
		WithConfigDirectory(tempDir).
		WithEnvironment("dev")

	_, err = appSettings.Load()
	if err == nil {
		t.Error("Load() should return error for invalid env config")
	}
}

func TestChaining(t *testing.T) {
	result := New[TestConfig]().
		WithArgs([]string{"--test"}).
		WithEnvVars([]string{"TEST=value"}).
		WithEnvironment("test").
		WithConfigDirectory("/tmp")

	if len(result.withArgs) == 0 {
		t.Error("Args should be set after chaining")
	}

	if len(result.withEnvVars) == 0 {
		t.Error("EnvVars should be set after chaining")
	}

	if result.withEnvironment == nil || *result.withEnvironment != "test" {
		t.Error("Environment should be set after chaining")
	}

	if result.withConfigDirectory == nil || *result.withConfigDirectory != "/tmp" {
		t.Error("ConfigDirectory should be set after chaining")
	}
}

func TestLoadConfigFile_ReadError(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	if err := os.Mkdir(configFile, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	appSettings := New[TestConfig]()
	configMap := make(map[string]interface{})

	err := appSettings.loadConfigFile(configFile, configMap)
	if err == nil {
		t.Error("loadConfigFile() should return error when trying to read a directory")
	}
}

func TestUnmarshalToType_MarshalError(t *testing.T) {
	appSettings := New[TestConfig]()

	configMap := map[string]interface{}{
		"channel": make(chan int),
	}

	_, err := appSettings.unmarshalToType(configMap)
	if err == nil {
		t.Error("unmarshalToType() should return error when marshal fails")
	}
}

func TestLoad_UnmarshalError(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	wrongConfig := map[string]interface{}{
		"port": "not-a-number",
	}

	data, err := json.Marshal(wrongConfig)
	if err != nil {
		t.Fatalf("Failed to marshal wrong config: %v", err)
	}
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	appSettings := New[TestConfig]().WithConfigDirectory(tempDir)

	_, err = appSettings.Load()
	if err == nil {
		t.Error("Load() should return error when unmarshal fails")
	}
}

type ComplexConfig struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
	Cache struct {
		Enabled bool   `json:"enabled"`
		TTL     int    `json:"ttl"`
		Type    string `json:"type"`
	} `json:"cache"`
	Features []string `json:"features"`
}

func TestLoad_ComplexConfig(t *testing.T) {
	tempDir := t.TempDir()

	baseConfig := map[string]interface{}{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "user",
			"password": "pass",
		},
		"cache": map[string]interface{}{
			"enabled": false,
			"ttl":     300,
			"type":    "memory",
		},
		"features": []string{"feature1", "feature2"},
	}

	baseData, err := json.Marshal(baseConfig)
	if err != nil {
		t.Fatalf("Failed to marshal base config: %v", err)
	}
	baseConfigFile := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(baseConfigFile, baseData, 0600); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	appSettings := New[ComplexConfig]().
		WithConfigDirectory(tempDir).
		WithEnvVars([]string{}).
		WithArgs([]string{})

	result, err := appSettings.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if result.Database.Host != "localhost" {
		t.Errorf("Expected database host 'localhost', got %s", result.Database.Host)
	}

	if result.Database.Port != 5432 {
		t.Errorf("Expected database port 5432, got %d", result.Database.Port)
	}

	if result.Cache.Enabled {
		t.Errorf("Expected cache enabled false, got %v", result.Cache.Enabled)
	}

	if len(result.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(result.Features))
	}
}
