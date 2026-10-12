//go:build !windows && !darwin && !linux

package main

import "errors"

func autoStartSupported() bool { return false }

func autoStartTarget() (string, error) {
	return "", errors.New("当前平台不支持开机自启")
}

func readAutoStart() (bool, string, error) { return false, "", nil }

func writeAutoStart(string, bool) error {
	return errors.New("当前平台不支持开机自启")
}
