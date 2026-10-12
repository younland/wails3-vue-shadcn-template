package main

import (
	"slices"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// 单实例：程序运行期间只允许一个实例（是否启用见 singleinstance_prod.go /
// singleinstance_dev.go）。重复启动时，新进程只把「显示主界面」的请求交给
// 已有实例，然后立刻退出，避免出现第二个托盘图标、第二个任务栏按钮或两份挂载状态。

// showWindowRequests 缓存显示请求；容量 1，连点多次会合并成一次。
var showWindowRequests = make(chan struct{}, 1)

// appStarted 在应用进入事件循环后关闭：窗口操作要等它之后才安全。
var appStarted = make(chan struct{})

var appStartedOnce sync.Once

// singleInstanceOptions 返回单实例配置；开发构建返回 nil，不限制多开。
func singleInstanceOptions(onSecondInstanceLaunch func(application.SecondInstanceData)) *application.SingleInstanceOptions {
	if !singleInstanceEnabled {
		return nil
	}
	return &application.SingleInstanceOptions{
		UniqueID:               appID,
		ExitCode:               0,
		OnSecondInstanceLaunch: onSecondInstanceLaunch,
	}
}

// hasTrayFlag 判断启动参数里是否带 --tray（跳过 argv[0]）。
func hasTrayFlag(args []string) bool {
	if len(args) < 2 {
		return false
	}
	return slices.Contains(args[1:], trayFlag)
}

// requestShowWindow 记下一次显示请求；可在事件循环启动前调用。
func requestShowWindow() {
	select {
	case showWindowRequests <- struct{}{}:
	default:
	}
}

// watchSecondInstanceShow 在应用启动后处理第二个实例的显示请求，须在 app.Run 之前调用。
func watchSecondInstanceShow(app *application.App, show func()) {
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		appStartedOnce.Do(func() { close(appStarted) })
	})
	go func() {
		for range showWindowRequests {
			<-appStarted
			show()
		}
	}()
}
