package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GenerateFilename создает имя файла по маске
// Маска: %name%-%Y%m%d%H%M%S
func GenerateFilename(mask, name string, t time.Time) string {
	result := mask
	result = strings.ReplaceAll(result, "%name%", name)
	result = strings.ReplaceAll(result, "%Y", t.Format("2006"))
	result = strings.ReplaceAll(result, "%m", t.Format("01"))
	result = strings.ReplaceAll(result, "%d", t.Format("02"))
	result = strings.ReplaceAll(result, "%H", t.Format("15"))
	result = strings.ReplaceAll(result, "%M", t.Format("04"))
	result = strings.ReplaceAll(result, "%S", t.Format("05"))
	return result
}

// ParseDateFromFilename извлекает дату из имени файла
// Формат: {name}-{YYYYMMDDHHmmss}.{ext}
func ParseDateFromFilename(filename string) (time.Time, error) {
	// Убираем расширение
	base := filename
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		base = filename[:idx]
	}

	// Ищем паттерн YYYYMMDDHHmmss в конце имени
	re := regexp.MustCompile(`(\d{14})$`)
	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("cannot parse date from filename: %s", filename)
	}

	dateStr := matches[1]
	t, err := time.Parse("20060102150405", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}

	return t, nil
}

// GetExtension возвращает расширение файла для типа сжатия
func GetExtension(compression string) string {
	switch compression {
	case "gzip":
		return ".gz"
	case "zip":
		return ".zip"
	case "tar":
		return ".tar"
	case "tar.gz":
		return ".tar.gz"
	default:
		return ""
	}
}

