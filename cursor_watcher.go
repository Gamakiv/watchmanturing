package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

// CursorWatcher представляет состояние отслеживания курсора
type CursorWatcher struct {
	isRunning   bool
	checkMillis int
	targetX     int
	targetY     int
}

// NewCursorWatcher создает новый экземпляр CursorWatcher
func NewCursorWatcher(checkMillis int, startX, startY int) *CursorWatcher {
	return &CursorWatcher{
		isRunning:   false,
		checkMillis: checkMillis,
		targetX:     startX,
		targetY:     startY,
	}
}

// Start запускает отслеживание курсора
func (cw *CursorWatcher) Start() {
	if cw.isRunning {
		fmt.Println("Курсор уже отслеживается!")
		return
	}

	cw.isRunning = true
	fmt.Println("🔥 Запущено отслеживание курсора!")

	go func() {
		// Устанавливаем курсор в начальную позицию
		robotgo.MoveMouse(cw.targetX, cw.targetY)
		fmt.Printf("Координаты курсора установлены: (%d, %d)\n", cw.targetX, cw.targetY)

		checkInterval := time.Duration(cw.checkMillis) * time.Millisecond

		for cw.isRunning {
			currentX, currentY := robotgo.GetMousePos()
			if currentX != cw.targetX || currentY != cw.targetY {
				fmt.Printf("🚨 Курсор сдвинулся! Новая позиция: (%d, %d)\n", currentX, currentY)
				cw.targetX, cw.targetY = currentX, currentY // Обновляем целевую позицию
				onCursorMoved(fmt.Sprintf("Курсор сдвинулся! Новая позиция: (%d, %d)", currentX, currentY))
			}
			time.Sleep(checkInterval)
		}
	}()
}

// Stop останавливает отслеживание курсора
func (cw *CursorWatcher) Stop() {
	if !cw.isRunning {
		fmt.Println("Курсор не отслеживается!")
		return
	}

	cw.isRunning = false
	fmt.Println("❌ Отслеживание курсора остановлено!")
}
