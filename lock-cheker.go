package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func isScreenLocked() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq LogonUI.exe")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "LogonUI.exe")
}

func main() {
	if isScreenLocked() {
		fmt.Println("Экран заблокирован")
		os.Exit(1) // Возвращаем код 1 при заблокированном экране
	} else {
		fmt.Println("Экран разблокирован")
		os.Exit(0) // Возвращаем код 0 при разблокированном экране
	}
}
