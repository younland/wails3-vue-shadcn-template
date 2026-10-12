//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const autoStartKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

func autoStartSupported() bool { return true }

// autoStartTarget 返回自启命令：当前可执行文件路径。
func autoStartTarget() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序路径失败: %w", err)
	}
	return filepath.Clean(exe), nil
}

// readAutoStart 从注册表读取开机自启项；值名使用 appID，避免与其它应用冲突。
func readAutoStart() (bool, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("打开注册表失败: %w", err)
	}
	defer key.Close()

	value, _, err := key.GetStringValue(appID)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("读取开机自启设置失败: %w", err)
	}
	return true, unquoteCommand(value), nil
}

func writeAutoStart(exePath string, enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, autoStartKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer key.Close()

	if enabled {
		// Run 项按命令行解析，含空格的路径必须加引号；--tray 让登录启动时静默进托盘
		if err := key.SetStringValue(appID, strconv.Quote(exePath)+" "+trayFlag); err != nil {
			return fmt.Errorf("写入开机自启设置失败: %w", err)
		}
		return nil
	}
	if err := key.DeleteValue(appID); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("移除开机自启设置失败: %w", err)
	}
	return nil
}

// unquoteCommand 取出命令中的程序路径（自动启动项带 --tray 参数）。
func unquoteCommand(value string) string {
	value = strings.TrimSpace(value)
	if unquoted, err := strconv.Unquote(value); err == nil {
		return unquoted
	}
	if strings.HasPrefix(value, `"`) {
		if end := strings.Index(value[1:], `"`); end >= 0 {
			if unquoted, err := strconv.Unquote(value[:end+2]); err == nil {
				return unquoted
			}
		}
	}
	return value
}
