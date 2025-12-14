package main

import (
	"fmt"
	"os"

	"backup-tool/backup"
	"backup-tool/config"
)

func main() {
	// Определяем путь к конфигу (по умолчанию config.yaml в текущей директории)
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	fmt.Printf("Loading configuration from %s...\n", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d backup(s) to process\n", len(cfg.Backups))

	executor := backup.NewExecutor(&cfg.Global)

	successCount := 0
	errorCount := 0

	for i, backupCfg := range cfg.Backups {
		fmt.Printf("\n[%d/%d] Processing backup: %s\n", i+1, len(cfg.Backups), backupCfg.Name)

		if err := executor.ExecuteBackup(&backupCfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing backup %s: %v\n", backupCfg.Name, err)
			errorCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", errorCount)

	if errorCount > 0 {
		os.Exit(1)
	}
}

