package tool

import (
    "runtime"
)

const (
    defaultStackSize = 4096
)

// getCurrentGoroutineStack 获取当前Goroutine的调用栈，便于排查panic异常
func GetCurrentGoroutineStack() string {
    var buf [defaultStackSize]byte
    n := runtime.Stack(buf[:], false)
    return string(buf[:n])
}
