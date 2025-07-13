package cmd

import (
	"bytes"
	"fmt"
	"gwnp/util"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)



type PHPServer struct {
	PhpCgiSpawnerFile string
	PhpPath           string
	Versions          []Server
}

func startPHP(PhpCgiSpawnerFile, PhpPath, path, port string) {
	// 启动PHP服务
	// TODO 使用php-cgi-spawner.exe启动PHP服务
	// 定义个切片，存储PHP的命令
	phpCgi := filepath.Join(PhpPath, path, "php-cgi.exe")
	phpIni := filepath.Join(PhpPath, path, "php.ini")
	arguments := []string{phpCgi + " -c " + phpIni, port, "1+4"}
	// fmt.Println(strings.Join(arguments, " "))
	cmd := exec.Command(PhpCgiSpawnerFile, arguments...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// phpCmd.Stdout = os.Stdout
	// phpCmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("failed to start php: %v\n", err)
		return
	}

}

func (p *PHPServer) Start() {
	var vs = []string{}
	util.NewUtilLog().Debug(fmt.Sprintf("Starting PHP versions %v", p.Versions))
	for _, php := range p.Versions {
		if php.Active == 1 {
			startPHP(p.PhpCgiSpawnerFile, p.PhpPath, php.Path, php.Port)
			util.NewUtilLog().Debug(fmt.Sprintf("Starting PHP %v", php))
			vs = append(vs, php.Version)
		}
	}

	fmt.Println("php", vs, "started successfully")
}

func (p *PHPServer) Stop() {
	// 停止PHP服务
	// fmt.Println("php stopping...")
	p.stopPcsw()
	// 执行命令并捕获输出
	cmd := exec.Command("taskkill", "/F", "/IM", "php-cgi.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Stdout = os.Stdout // 将命令的输出重定向到标准输出
	// cmd.Stderr = os.Stderr // 将命令的错误输出重定向到标准错误输出
	if err := cmd.Run(); err != nil {
		fmt.Printf("failed to stop php: %v\n", err)
		return
	}
	fmt.Println("php stopped")
}

func (p *PHPServer) stopPcsw() {
	// 执行命令并捕获输出
	cmd := exec.Command("taskkill", "/F", "/IM", "php-cgi-spawner.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Stdout = os.Stdout // 将命令的输出重定向到标准输出
	// cmd.Stderr = os.Stderr // 将命令的错误输出重定向到标准错误输出
	if err := cmd.Run(); err != nil {
		fmt.Printf("failed to stop php-cgi-spawner: %v\n", err)
		return
	}
}

func (p *PHPServer) Status() bool {
	// 执行命令并捕获输出
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq php-cgi.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// 获取命令输出
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("failed to check php: %v \n", err)
		return false
	}
	outString := out.String()
	// fmt.Println(outString)
	// 检查输出中是否包含Nginx
	if strings.Contains(outString, "php") {
		log.Println("php is running")
		return true
	} else {
		log.Println("php is not running")
		return false
	}
}
