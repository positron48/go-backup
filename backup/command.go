package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecuteCommand выполняет команду и проверяет наличие output_file
func ExecuteCommand(command string, outputFile string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("empty command")
	}

	// Выполняем команду через shell для поддержки многострочных команд и пайпов
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	// Проверяем, что output_file существует
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return fmt.Errorf("output file does not exist after command execution: %s", outputFile)
	}

	return nil
}

