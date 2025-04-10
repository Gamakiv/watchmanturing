package main

import (
	"fmt"
	"syscall"
	"time"
)

// Функция для блокировки экрана
func lockScreen() error {
	// Загружаем user32.dll
	user32 := syscall.NewLazyDLL("user32.dll")
	// Получаем функцию LockWorkStation
	lockWorkStation := user32.NewProc("LockWorkStation")

	// Вызываем функцию
	ret, _, err := lockWorkStation.Call()
	if ret == 0 {
		return fmt.Errorf("ошибка при блокировке экрана: %v", err)
	}

	// Даём время для блокировки перед выходом
	time.Sleep(1 * time.Second)
	return nil
}
