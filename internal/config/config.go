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
