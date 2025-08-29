package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadConfig() (Config, error) {
	var config Config

	// Try multiple config file locations
	configPaths := []string{
		"docs-config.json",
		".docs-config.json",
		".config/docs.json",
		"kolosys-docs.json", // New centralized naming
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return config, fmt.Errorf("no configuration file found. Please create docs-config.json")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// Apply defaults for missing values
	if config.Docs.RootDir == "" {
		config.Docs.RootDir = "."
	}
	if config.Docs.DocsDir == "" {
		config.Docs.DocsDir = "docs"
	}
	if config.Repository.ImportPath == "" {
		config.Repository.ImportPath = fmt.Sprintf("github.com/%s/%s",
			config.Repository.Owner, config.Repository.Name)
	}

	fmt.Printf("📄 Loaded configuration from %s\n", configFile)
	return config, nil
}
