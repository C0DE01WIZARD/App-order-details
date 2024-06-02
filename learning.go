package main

import (
    "fmt"
    "time"
)

// Функция, которая будет выполняться в горутине
func say(s string) {
    for i := 0; i < 5; i++ {
        time.Sleep(1000 * time.Millisecond)
        fmt.Println(s)
    }
}

func main() {
    // Запуск функции say в горутине
    go say("world")
    // Основная горутина продолжает выполняться
    say("hello")
}
