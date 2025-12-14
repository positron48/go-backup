package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyDirectory копирует директорию с поддержкой exclude_patterns
func CopyDirectory(source, destination string, excludePatterns []string) error {
	// Создаем целевую директорию
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Проверяем exclude patterns
		if shouldExclude(relPath, excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(destination, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath, info.Mode())
	})
}

func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue
		}

		if matched {
			return true
		}

		// Также проверяем, начинается ли путь с паттерна (для директорий)
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}

	return false
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

