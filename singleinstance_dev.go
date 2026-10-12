//go:build !production

package main

// 开发构建（wails3 dev、task build DEV=true）不启用单实例：
// 重新构建出的新进程不该被上一次开发运行的实例拦下。
const singleInstanceEnabled = false
