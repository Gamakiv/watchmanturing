package main

import (
	"os/exec"
	"strings"
)

// Проверяет, заблокирован ли экран
func isScreenLocked() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq LogonUI.exe")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "LogonUI.exe")
}

// Возвращает строковое представление состояния экрана
func getScreenStatus() string {
	if isScreenLocked() {
		return "Экран заблокирован"
	}
	return "Экран разблокирован"
}
