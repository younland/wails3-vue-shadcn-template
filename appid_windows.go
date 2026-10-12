//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// init 在创建任何窗口之前显式设置进程 AUMID。
// 任务栏按 AUMID 分组：多个窗口要像 VS Code 多窗口那样收在同一个任务栏按钮下，
// 必须共享同一标识；否则 Windows 会按可执行文件推断身份，可能把窗口拆成多个按钮。
// 该标识与通知使用的 AUMID 一致，任务栏名称/图标也由此登记信息决定。
func init() {
	setProcessAppUserModelID(appID)
}

func setProcessAppUserModelID(id string) {
	ptr, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		return
	}
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("SetCurrentProcessExplicitAppUserModelID")
	_, _, _ = proc.Call(uintptr(unsafe.Pointer(ptr)))
}
