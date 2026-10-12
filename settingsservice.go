package main

import "errors"

// AutoStartStatus 描述开机自启的当前状态。
type AutoStartStatus struct {
	Enabled   bool   `json:"enabled"`
	Supported bool   `json:"supported"`
	Target    string `json:"target"`
}

// SettingsService 提供应用级设置的读写。
type SettingsService struct{}

// AutoStart 返回当前的开机自启状态。
func (s *SettingsService) AutoStart() (AutoStartStatus, error) {
	if !autoStartSupported() {
		return AutoStartStatus{}, nil
	}
	enabled, target, err := readAutoStart()
	if err != nil {
		return AutoStartStatus{}, err
	}
	return AutoStartStatus{Enabled: enabled, Supported: true, Target: target}, nil
}

// SetAutoStart 打开或关闭开机自启，自启项指向当前可执行文件。
func (s *SettingsService) SetAutoStart(enabled bool) error {
	if !autoStartSupported() {
		return errors.New("当前平台不支持开机自启")
	}
	target, err := autoStartTarget()
	if err != nil {
		return err
	}
	return writeAutoStart(target, enabled)
}

// Version 返回当前应用版本（发布构建由 CI 通过 -ldflags 注入 appVersion）。
func (s *SettingsService) Version() string {
	return appVersion
}
