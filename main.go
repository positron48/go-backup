package main

import (
	"fmt"
	"os"

	"backup-tool/backup"
	"backup-tool/config"
	"backup-tool/utils"
)

func main() {
	// Определяем путь к конфигу (по умолчанию config.yaml в текущей директории)
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	utils.PrintHeader("Loading configuration from %s...", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		utils.PrintError("Error loading config: %v", err)
		os.Exit(1)
	}

	utils.PrintHeader("Found %d backup(s) to process", len(cfg.Backups))

	executor := backup.NewExecutor(&cfg.Global)

	successCount := 0
	errorCount := 0

	for i, backupCfg := range cfg.Backups {
		utils.PrintHeaderf("\n[%d/%d] Processing backup: %s\n", i+1, len(cfg.Backups), backupCfg.Name)

		if err := executor.ExecuteBackup(&backupCfg); err != nil {
			utils.PrintError("Error executing backup %s: %v", backupCfg.Name, err)
			errorCount++
			continue
		}

		successCount++
	}

	utils.PrintHeader("\n=== Summary ===")
	if successCount > 0 {
		utils.PrintSuccess("Successful: %d", successCount)
	} else {
		fmt.Printf("Successful: %d\n", successCount)
	}
	if errorCount > 0 {
		utils.PrintError("Failed: %d", errorCount)
	} else {
		fmt.Printf("Failed: %d\n", errorCount)
	}

	if errorCount > 0 {
		os.Exit(1)
	}
}

