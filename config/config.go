package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type RetentionPolicy struct {
	Daily   int `yaml:"daily"`
	Weekly  int `yaml:"weekly"`
	Monthly int `yaml:"monthly"`
	Yearly  int `yaml:"yearly"`
}

type GlobalConfig struct {
	BackupDir          string          `yaml:"backup_dir"`
	Retention          RetentionPolicy `yaml:"retention"`
	FilenameMask       string          `yaml:"filename_mask"`
	DefaultCompression string          `yaml:"default_compression"`
	PreHooks           []string        `yaml:"pre_hooks"`
	PostHooks          []string        `yaml:"post_hooks"`
	IncludeDir         string          `yaml:"include_dir"`
}

type BackupConfig struct {
	Name            string           `yaml:"name"`
	Subdirectory    string           `yaml:"subdirectory"`
	SourceDir       string           `yaml:"source_dir"`
	Command         string           `yaml:"command"`
	OutputFile      string           `yaml:"output_file"`
	Compression     string           `yaml:"compression"`
	ExcludePatterns []string         `yaml:"exclude_patterns"`
	Retention       *RetentionPolicy `yaml:"retention"`
	PreHooks        []string         `yaml:"pre_hooks"`
	PostHooks       []string         `yaml:"post_hooks"`
}

type Config struct {
	Global  GlobalConfig   `yaml:"global"`
	Backups []BackupConfig `yaml:"backups"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Загружаем бэкапы из include_dir
	if config.Global.IncludeDir != "" {
		backups, err := loadBackupsFromDir(config.Global.IncludeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load backups from include_dir: %w", err)
		}
		config.Backups = append(config.Backups, backups...)
	}

	// Валидация
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func loadBackupsFromDir(dir string) ([]BackupConfig, error) {
	var backups []BackupConfig

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".yaml") && !strings.HasSuffix(strings.ToLower(name), ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		var backup BackupConfig
		if err := yaml.Unmarshal(data, &backup); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

func validateConfig(config *Config) error {
	if config.Global.BackupDir == "" {
		return fmt.Errorf("backup_dir is required")
	}

	if config.Global.FilenameMask == "" {
		return fmt.Errorf("filename_mask is required")
	}

	if config.Global.DefaultCompression == "" {
		config.Global.DefaultCompression = "none"
	}

	for i, backup := range config.Backups {
		if backup.Name == "" {
			return fmt.Errorf("backup[%d]: name is required", i)
		}

		if backup.Subdirectory == "" {
			return fmt.Errorf("backup[%d]: subdirectory is required", i)
		}

		// Должен быть либо source_dir, либо (command + output_file)
		hasSourceDir := backup.SourceDir != ""
		hasCommand := backup.Command != "" && backup.OutputFile != ""

		if !hasSourceDir && !hasCommand {
			return fmt.Errorf("backup[%d]: must have either source_dir or (command + output_file)", i)
		}

		if hasSourceDir && hasCommand {
			return fmt.Errorf("backup[%d]: cannot have both source_dir and command", i)
		}
	}

	return nil
}

