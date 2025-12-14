package utils

import (
	"fmt"
	"os"
)

const (
	// ANSI color codes
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorOrange = "\033[33m"
	ColorYellow = "\033[33m"
)

// PrintSuccess выводит успешное сообщение зеленым цветом
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("%s%s%s\n", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

// PrintError выводит сообщение об ошибке красным цветом
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s%s\n", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

// PrintHeader выводит заголовок оранжевым цветом
func PrintHeader(format string, args ...interface{}) {
	fmt.Printf("%s%s%s\n", ColorOrange, fmt.Sprintf(format, args...), ColorReset)
}

// PrintSuccessf выводит успешное сообщение зеленым цветом (аналог Printf)
func PrintSuccessf(format string, args ...interface{}) {
	fmt.Printf("%s%s%s", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

// PrintErrorf выводит сообщение об ошибке красным цветом (аналог Printf)
func PrintErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s%s", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

// PrintHeaderf выводит заголовок оранжевым цветом (аналог Printf)
func PrintHeaderf(format string, args ...interface{}) {
	fmt.Printf("%s%s%s", ColorOrange, fmt.Sprintf(format, args...), ColorReset)
}

