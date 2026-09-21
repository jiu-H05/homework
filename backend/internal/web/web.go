// Package web 通过 go:embed 把前端静态产物（dist/，相当于 Vue 的打包目录）
// 编译进后端二进制，实现“一个可执行文件、页面与 API 同一端口”。
package web

import (
	"embed"
	"io/fs"
)

// Dist 是嵌入的前端构建产物根目录。
//
//go:embed all:dist
var Dist embed.FS

// Static 返回以 dist 为根的只读文件系统，供 HTTP 文件服务使用。
func Static() fs.FS {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic(err) // 路径固定为编译期常量，理论上不会发生
	}
	return sub
}
