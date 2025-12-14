package hooks

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunHooks(hooks []string) error {
	for _, hook := range hooks {
		hook = strings.TrimSpace(hook)
		if hook == "" {
			continue
		}

		// Разбиваем команду на части
		parts := strings.Fields(hook)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Логируем ошибку, но не прерываем процесс
			fmt.Printf("Hook failed: %s\nOutput: %s\nError: %v\n", hook, string(output), err)
			continue
		}

		if len(output) > 0 {
			fmt.Printf("Hook output: %s\n", string(output))
		}
	}

	return nil
}

