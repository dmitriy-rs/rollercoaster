package config

import (
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/spf13/viper"
)

type Config struct {
	DefaultJSManager  string
	AutoSelectClosest bool
}

type ConfigItem struct {
	Key         string
	Label       string
	Description string
	ItemType    string // "boolean", "select"
	Value       interface{}
	Options     []string // for select type
}

func LoadConfig() *Config {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err := createConfig()
			if err != nil {
				logger.Error("Error creating config", err)
				return nil
			}
		} else {
			logger.Error("Error loading config", err)
			return nil
		}
	}

	logger.Debug("Config loaded successfully")

	defaultJSManager := validateDefaultJSManager(viper.GetString("DefaultJSManager"))
	enableDefaultJSManager := viper.GetBool("EnableDefaultJSManager")
	autoSelectClosest := viper.GetBool("AutoSelectClosest")

	if enableDefaultJSManager {
		return &Config{
			DefaultJSManager:  defaultJSManager,
			AutoSelectClosest: autoSelectClosest,
		}
	}

	return &Config{
		DefaultJSManager:  "",
		AutoSelectClosest: autoSelectClosest,
	}
}

func createConfig() error {
	configDir := path.Join(os.Getenv("HOME"), ".rollercoaster")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			logger.Error("Error creating config directory", err)
			return err
		}
	}

	viper.SetConfigFile(path.Join(os.Getenv("HOME"), ".rollercoaster", "config.toml"))
	viper.SetConfigType("toml")

	viper.SetDefault("EnableDefaultJSManager", false)
	viper.SetDefault("DefaultJSManager", "npm")
	viper.SetDefault("AutoSelectClosest", true)

	err := viper.WriteConfig()
	if err != nil {
		return err
	}

	logger.Debug("Config created successfully")
	return nil
}

func validateDefaultJSManager(defaultJSManager string) string {
	switch defaultJSManager {
	case "npm":
		return "npm"
	case "yarn":
		return "yarn"
	case "pnpm":
		return "pnpm"
	case "bun":
		return "bun"
	case "deno":
		return "deno"
	}
	logger.Warning("Invalid default JS manager: " + defaultJSManager + ". Allowed values are: npm, yarn, pnpm, bun, deno")
	return ""
}

// GetAllConfigItems returns all available configuration items with their current values
func GetAllConfigItems() []ConfigItem {
	// Ensure config is loaded
	loadConfigFile()

	return []ConfigItem{
		{
			Key:         "EnableDefaultJSManager",
			Label:       "Enable Default JS Manager",
			Description: "Enable the default JavaScript package manager for projects",
			ItemType:    "boolean",
			Value:       viper.GetBool("EnableDefaultJSManager"),
		},
		{
			Key:         "DefaultJSManager",
			Label:       "Default JS Manager",
			Description: "Choose the default JavaScript package manager (npm, yarn, pnpm, bun, deno)",
			ItemType:    "select",
			Value:       viper.GetString("DefaultJSManager"),
			Options:     []string{"npm", "yarn", "pnpm", "bun", "deno"},
		},
		{
			Key:         "AutoSelectClosest",
			Label:       "Auto Select Closest Match",
			Description: "Automatically select the closest matching task when multiple matches are found",
			ItemType:    "boolean",
			Value:       viper.GetBool("AutoSelectClosest"),
		},
	}
}

// GetConfigValue returns the value of a specific configuration key
func GetConfigValue(key string) interface{} {
	loadConfigFile()
	return viper.Get(key)
}

// SetConfigValue sets a configuration value
func SetConfigValue(key string, value interface{}) {
	viper.Set(key, value)
}

// SaveConfig saves the current configuration to file
func SaveConfig() error {
	// Ensure config file exists
	configDir := path.Join(os.Getenv("HOME"), ".rollercoaster")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			logger.Error("Error creating config directory", err)
			return err
		}
	}

	viper.SetConfigFile(path.Join(os.Getenv("HOME"), ".rollercoaster", "config.toml"))
	viper.SetConfigType("toml")

	// Write config to file
	err := viper.WriteConfig()
	if err != nil {
		logger.Error("Error saving config", err)
		return err
	}

	logger.Debug("Config saved successfully")
	return nil
}

// UpdateConfigItems updates multiple configuration items at once
func UpdateConfigItems(items []ConfigItem) error {
	for _, item := range items {
		viper.Set(item.Key, item.Value)
	}
	return SaveConfig()
}

// loadConfigFile ensures the configuration file is loaded
func loadConfigFile() {
	viper.SetConfigFile(path.Join(os.Getenv("HOME"), ".rollercoaster", "config.toml"))
	viper.SetConfigType("toml")

	// Set defaults in case config file doesn't exist
	viper.SetDefault("EnableDefaultJSManager", false)
	viper.SetDefault("DefaultJSManager", "npm")
	viper.SetDefault("AutoSelectClosest", true)

	// Try to read config, create if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create it
			createConfig()
		}
	}
}
