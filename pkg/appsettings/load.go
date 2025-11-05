// Package appsettings provides a generic application settings loader with layered configuration sources.
package appsettings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// AppSettings is a generic configuration loader that supports layered sources:
// command line arguments, environment variables, environment-specific config files, and base config files.
type AppSettings[T any] struct {
	withArgs            []string
	withEnvVars         []string
	withEnvironment     *string
	withConfigDirectory *string
}

// New creates a new AppSettings instance for the given config type.
func New[T any]() *AppSettings[T] {
	return &AppSettings[T]{
		withArgs:            nil,
		withEnvVars:         nil,
		withEnvironment:     nil,
		withConfigDirectory: nil,
	}
}

// Load loads the configuration in the following priority order:
// Args > EnvVars > ConfigFile.env.json > ConfigFile.json.
// It returns a pointer to the populated config struct of type T.
func (a *AppSettings[T]) Load() (*T, error) {
	configMap := make(map[string]interface{})

	// Get working directory or config directory
	configDir, err := a.getConfigDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	// Load base config file
	baseConfigPath := filepath.Join(*configDir, "config.json")
	if err := a.loadConfigFile(baseConfigPath, configMap); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// Load environment-specific config file
	if a.withEnvironment != nil {
		envConfigPath := filepath.Join(*configDir, fmt.Sprintf("config.%s.json", *a.withEnvironment))
		if err := a.loadConfigFile(envConfigPath, configMap); err != nil {
			return nil, fmt.Errorf("failed to load env config: %w", err)
		}
	}

	// Overlay environment variables
	if err := a.loadEnvVars(configMap); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// Overlay command line arguments
	if err := a.loadArgs(configMap); err != nil {
		return nil, fmt.Errorf("failed to load args: %w", err)
	}

	// Unmarshal map into T and return
	result, err := a.unmarshalToType(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return result, nil
}

// getWD returns the directory of the running executable.
func (a *AppSettings[T]) getWD() (*string, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exPath := filepath.Dir(ex)
	return &exPath, nil
}

// WithArgs sets the command line arguments to use for configuration.
func (a *AppSettings[T]) WithArgs(args []string) *AppSettings[T] {
	a.withArgs = args
	return a
}

// WithEnvVars sets the environment variables to use for configuration.
func (a *AppSettings[T]) WithEnvVars(env []string) *AppSettings[T] {
	a.withEnvVars = env
	return a
}

// WithEnvironment sets the environment name (e.g., "dev", "prod") for environment-specific config file loading.
func (a *AppSettings[T]) WithEnvironment(environment string) *AppSettings[T] {
	a.withEnvironment = &environment
	return a
}

// WithConfigDirectory sets the directory to look for config files.
func (a *AppSettings[T]) WithConfigDirectory(configDirectory string) *AppSettings[T] {
	a.withConfigDirectory = &configDirectory
	return a
}

// getConfigDirectory returns the config directory, falling back to the executable directory if not set.
func (a *AppSettings[T]) getConfigDirectory() (*string, error) {
	if a.withConfigDirectory != nil {
		return a.withConfigDirectory, nil
	}
	return a.getWD()
}

// loadConfigFile loads a JSON config file and merges its values into configMap.
// If the file does not exist, it is silently ignored.
func (a *AppSettings[T]) loadConfigFile(filePath string, configMap map[string]interface{}) error {
	//nolint:gosec // filePath is constructed from trusted config directory and filename
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Config file is optional
		}
		return err
	}

	var fileConfig map[string]interface{}
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return err
	}

	// Merge into configMap
	for key, value := range fileConfig {
		configMap[key] = value
	}

	return nil
}

// loadEnvVars overlays environment variables into configMap, converting values to appropriate types.
func (a *AppSettings[T]) loadEnvVars(configMap map[string]interface{}) error {
	if a.withEnvVars == nil {
		return nil
	}

	for _, envVar := range a.withEnvVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		// Convert value to appropriate type if possible
		configMap[key] = a.parseValue(value)
	}

	return nil
}

// loadArgs overlays command line arguments into configMap, converting values to appropriate types.
// Supports --key value and --flag formats.
func (a *AppSettings[T]) loadArgs(configMap map[string]interface{}) error {
	if a.withArgs == nil {
		return nil
	}

	for i, arg := range a.withArgs {
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			key = strings.ToLower(key)

			// Check if there's a value after this argument
			if i+1 < len(a.withArgs) && !strings.HasPrefix(a.withArgs[i+1], "--") {
				value := a.withArgs[i+1]
				configMap[key] = a.parseValue(value)
			} else {
				configMap[key] = true // Flag without value
			}
		}
	}

	return nil
}

// parseValue attempts to convert a string to bool, int, float, or returns the original string.
func (a *AppSettings[T]) parseValue(value string) interface{} {
	// Try to parse as bool
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try to parse as int
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Return as string
	return value
}

// unmarshalToType marshals configMap to JSON and unmarshals it into type T.
func (a *AppSettings[T]) unmarshalToType(configMap map[string]interface{}) (*T, error) {
	jsonData, err := json.Marshal(configMap)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
