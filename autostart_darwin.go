//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// autoStartLabel 使用 appID，与 LaunchAgents 的命名惯例一致。
var autoStartLabel = appID

func autoStartSupported() bool { return true }

func autoStartTarget() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序路径失败: %w", err)
	}
	return filepath.Clean(exe), nil
}

func autoStartPlist() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", autoStartLabel+".plist"), nil
}

func readAutoStart() (bool, string, error) {
	path, err := autoStartPlist()
	if err != nil {
		return false, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("读取开机自启配置失败: %w", err)
	}
	return true, plistTarget(string(data)), nil
}

func writeAutoStart(exePath string, enabled bool) error {
	path, err := autoStartPlist()
	if err != nil {
		return err
	}
	if !enabled {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("移除开机自启配置失败: %w", err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建 LaunchAgents 目录失败: %w", err)
	}
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, autoStartLabel, xmlEscape(exePath), trayFlag)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入开机自启配置失败: %w", err)
	}
	return nil
}

// plistTarget 取出 ProgramArguments 中的第一个条目。
func plistTarget(content string) string {
	marker := "<key>ProgramArguments</key>"
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(marker):]
	start := strings.Index(rest, "<string>")
	end := strings.Index(rest, "</string>")
	if start < 0 || end < start {
		return ""
	}
	return xmlUnescape(rest[start+len("<string>") : end])
}

func xmlEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(s)
}

func xmlUnescape(s string) string {
	replacer := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">")
	return replacer.Replace(s)
}
