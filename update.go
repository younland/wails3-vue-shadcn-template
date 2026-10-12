package main

// appVersion 是当前版本。
// 发布构建由 CI 按 tag 注入（-ldflags "-X main.appVersion=x.y.z"），
// 这里是没有注入时的默认版本；发布前请与 build/config.yml 的 version 保持一致。
var appVersion = "0.1.0"
