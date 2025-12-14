package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"goback/backup"
	"goback/config"
	"goback/hooks"
	"goback/utils"
)

func main() {
	// Парсим флаги командной строки
	var configPath string
	var backupNames flagArray
	var skipGlobalPreHooks bool
	var skipGlobalPostHooks bool

	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file (short)")
	flag.Var(&backupNames, "backup", "Name of backup to run (can be specified multiple times)")
	flag.Var(&backupNames, "b", "Name of backup to run (short, can be specified multiple times)")
	flag.BoolVar(&skipGlobalPreHooks, "skip-global-pre-hooks", false, "Skip global pre-hooks execution")
	flag.BoolVar(&skipGlobalPreHooks, "skip-pre-hooks", false, "Skip global pre-hooks execution (short)")
	flag.BoolVar(&skipGlobalPostHooks, "skip-global-post-hooks", false, "Skip global post-hooks execution")
	flag.BoolVar(&skipGlobalPostHooks, "skip-post-hooks", false, "Skip global post-hooks execution (short)")

	flag.Parse()

	// Обрабатываем позиционные аргументы для обратной совместимости
	// Формат: ./goback [config.yaml] [backup1] [backup2] ...
	args := flag.Args()
	if len(args) > 0 {
		// Первый аргумент может быть конфигом, если заканчивается на .yaml/.yml
		firstArg := args[0]
		if strings.HasSuffix(strings.ToLower(firstArg), ".yaml") || strings.HasSuffix(strings.ToLower(firstArg), ".yml") {
			configPath = firstArg
			args = args[1:] // Остальные аргументы - имена бэкапов
		}

		// Все оставшиеся аргументы - имена бэкапов
		backupNames = append(backupNames, args...)
	}

	utils.PrintHeader("Loading configuration from %s...", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		utils.PrintError("Error loading config: %v", err)
		os.Exit(1)
	}

	// Фильтруем бэкапы по указанным именам
	backupsToProcess := cfg.Backups
	if len(backupNames) > 0 {
		backupMap := make(map[string]*config.BackupConfig)
		for i := range cfg.Backups {
			backupMap[cfg.Backups[i].Name] = &cfg.Backups[i]
		}

		backupsToProcess = []config.BackupConfig{}
		var notFound []string
		for _, name := range backupNames {
			if backup, exists := backupMap[name]; exists {
				backupsToProcess = append(backupsToProcess, *backup)
			} else {
				notFound = append(notFound, name)
			}
		}

		if len(notFound) > 0 {
			utils.PrintError("Backup(s) not found: %s", strings.Join(notFound, ", "))
			os.Exit(1)
		}
	}

	utils.PrintHeader("Found %d backup(s) to process", len(backupsToProcess))

	// Выполняем глобальные pre-hooks перед всеми бэкапами
	if !skipGlobalPreHooks && len(cfg.Global.PreHooks) > 0 {
		utils.PrintHeader("Running global pre-hooks...")
		if err := hooks.RunHooks(cfg.Global.PreHooks); err != nil {
			fmt.Printf("Warning: global pre-hooks completed with errors\n")
		}
	}

	executor := backup.NewExecutor(&cfg.Global)

	successCount := 0
	errorCount := 0

	for i, backupCfg := range backupsToProcess {
		utils.PrintHeaderf("\n[%d/%d] Processing backup: %s\n", i+1, len(backupsToProcess), backupCfg.Name)

		if err := executor.ExecuteBackup(&backupCfg); err != nil {
			utils.PrintError("Error executing backup %s: %v", backupCfg.Name, err)
			errorCount++
			continue
		}

		successCount++
	}

	// Выполняем глобальные post-hooks после всех бэкапов
	if !skipGlobalPostHooks && len(cfg.Global.PostHooks) > 0 {
		utils.PrintHeader("\nRunning global post-hooks...")
		if err := hooks.RunHooks(cfg.Global.PostHooks); err != nil {
			fmt.Printf("Warning: global post-hooks completed with errors\n")
		}
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

// flagArray для поддержки множественных значений флага
type flagArray []string

func (f *flagArray) String() string {
	return strings.Join(*f, ", ")
}

func (f *flagArray) Set(value string) error {
	*f = append(*f, value)
	return nil
}

