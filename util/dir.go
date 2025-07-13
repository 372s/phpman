package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type UtilDirectory struct {}

func (u *UtilDirectory) RootPath() string {
	// 获取可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("获取当前工作目录失败:", err)
		return ""
	}
	// fmt.Println("当前可执行文件路径:", exePath)
	var exeDir string
	// 判断是否是 go run 模式
	if strings.Contains(exePath, os.TempDir()) || strings.Contains(exePath, "go-build") || strings.Contains(exePath, "main.exe") {
		// 开发模式 - 使用工作目录
		exeDir, err = os.Getwd()
		// fmt.Println("开发模式", "当前工作目录:", exeDir)
		if err != nil {
			fmt.Println("获取当前工作目录失败:", err)
			return ""
		}
	} else {
		// 生产模式 - 使用可执行文件所在目录
		// 获取可执行文件所在目录
		exeDir = filepath.Dir(exePath)
	}
	return exeDir
}