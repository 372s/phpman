package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type NginxServer struct {
	NginxPath string
	VersionPath string
}


func (n *NginxServer) Start() {
	path := filepath.Join(n.NginxPath, n.VersionPath)
	// fmt.Println("nginx starting...")
	// 定义个切片，存储命令
	params := []string{"-p", path}
	exe := filepath.Join(path, "nginx.exe")
	cmd := exec.Command(exe, params...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("failed to start nginx: %v\n", err)
		return
	}
	fmt.Println("nginx started successfully")
}

func (n *NginxServer) Stop() {
	// 停止PHP服务
	// fmt.Println("nginx stopping...")
	// 执行命令并捕获输出
	cmd := exec.Command("taskkill", "/F", "/IM", "nginx.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Stdout = os.Stdout // 将命令的输出重定向到标准输出
	// cmd.Stderr = os.Stderr // 将命令的错误输出重定向到标准错误输出
	if err := cmd.Run(); err != nil {
		// fmt.Printf("停止nginx失败: %v\n", err)
		return
	}
	fmt.Println("nginx stopped")
}

func (n *NginxServer) Status() bool {
	// 执行命令并捕获输出
	// 如何执行命令时不弹出弹窗
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq nginx.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Stdout = os.Stdout // 将命令的输出重定向到标准输出
	// cmd.Stderr = os.Stderr // 将命令的错误输出重定向到标准错误输出
	// if err := cmd.Run(); err != nil {
	// 	fmt.Printf("获取nginx状态失败: %v\n", err)
	// 	return "unknown", err
	// }
	out, err := cmd.Output()
	if err != nil {
		// 获取Nginx状态失败，转为英文描述文字
		// fmt.Printf("获取Nginx状态失败: %v", err)
		fmt.Printf("failed to nginx status error : %v\n", err)
		return false
	}
	
	s := string(out)
	if strings.Contains(s, "nginx.exe") {
		log.Println("nginx is running")
		return true
	}
	log.Println("nginx is not running")
	return false
}
