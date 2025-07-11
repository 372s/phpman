//go:build windows
// +build windows

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gwnp/cmd"
	"os"
	"path/filepath"
	"strings"

	"github.com/getlantern/systray"
)

type PhpVersions map[string]bool

type PhpVersionList []cmd.PhpVersionEntry

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	root := initRootPath()
	configFile := filepath.Join(root, "settings.json")
	nginxPath := filepath.Join(root, "servers", "nginx")
	phpPath := filepath.Join(root, "servers", "php")
	iconPath := filepath.Join(root, "icon")
	nginxVersion := "nginx-1.22.1"
	phpCgiSpawnerFile := filepath.Join(root, "bin", "php-cgi-spawner", "php-cgi-spawner.exe")

	// 判断是否有settings.json文件
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		phpList := readPHPDIR(phpPath)
		writeToJson(phpList, configFile)
	}

	// 读取settings.json文件，加到phpServer.Versions中
	phpVersions := readJson(configFile)

	// 创建server
	phpServer := &cmd.PHPServer{PhpCgiSpawnerFile: phpCgiSpawnerFile, PhpPath: phpPath, Versions: phpVersions}
	nginxServer := &cmd.NginxServer{NginxPath: nginxPath, VersionPath: nginxVersion}

	// 配置菜单
	systray.SetTitle("GNP")
	systray.SetTooltip("GNP")
	setTempIcon := func() {
		if phpServer.Status() && nginxServer.Status() {
			startIcon, _ := os.ReadFile(filepath.Join(iconPath, "start.ico"))
			systray.SetTemplateIcon(startIcon, startIcon)
		} else if nginxServer.Status() || phpServer.Status() {
			someoneIco, _ := os.ReadFile(filepath.Join(iconPath, "someone.ico"))
			systray.SetTemplateIcon(someoneIco, someoneIco)
		} else {
			stopIco, _ := os.ReadFile(filepath.Join(iconPath, "stop.ico"))
			systray.SetTemplateIcon(stopIco, stopIco)
		}
	}
	go func() {
		setTempIcon()
	}()

	mPHP := systray.AddMenuItem("选择启用的PHP版本(可多选)", "")
	mPHP.Disable()
	// 创建php版本菜单
	for k, php := range phpVersions {
		// fmt.Println(k, php)
		tray := systray.AddMenuItemCheckbox(php.Name, php.Name, php.Active)
		go func(tray *systray.MenuItem, k int, php cmd.PhpVersionEntry, configFile string) {
			for {
				<-tray.ClickedCh
				fmt.Println("Clicked on", tray, php)
				if tray.Checked() {
					tray.Uncheck()
					phpVersions[k].Active = false
				} else {
					tray.Check()
					// 如何覆盖phpVersions数据
					phpVersions[k].Active = true
				}
				writeToJson(phpVersions, configFile)
			}
		}(tray, k, php, configFile)
	}

	// 添加一个分隔符
	systray.AddSeparator()

	// 添加一个菜单项
	server := systray.AddMenuItem("启动服务", "启动服务")
	server.Disable()

	all := systray.AddMenuItem("所有服务", "")
	allStart := all.AddSubMenuItem("start", "start")
	allStop := all.AddSubMenuItem("stop", "stop")
	allRestart := all.AddSubMenuItem("restart", "restart")

	// 检查php是否启动h状态

	php := systray.AddMenuItemCheckbox("php", "php", phpServer.Status())
	phpStart := php.AddSubMenuItem("start", "start")
	phpStop := php.AddSubMenuItem("stop", "stop")
	phpRestart := php.AddSubMenuItem("restart", "restart")

	// 检查nginx是否启动
	nginx := systray.AddMenuItemCheckbox("nginx", "nginx", nginxServer.Status())
	nginxStart := nginx.AddSubMenuItem("start", "start")
	nginxStop := nginx.AddSubMenuItem("stop", "stop")
	nginxRestart := nginx.AddSubMenuItem("restart", "restart")

	// 退出
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-allStart.ClickedCh:
				fmt.Println("---- All start ---- ")
				phpServer.Start()
				nginxServer.Start()
				php.Check()
				nginx.Check()
			case <-allStop.ClickedCh:
				fmt.Println("---- All Stop ---- ")
				phpServer.Stop()
				nginxServer.Stop()
				php.Uncheck()
				nginx.Uncheck()
			case <-allRestart.ClickedCh:
				fmt.Println("---- All Restart ---- ")
				phpServer.Stop()
				nginxServer.Stop()
				phpServer.Start()
				nginxServer.Start()
				php.Check()
				nginx.Check()
			case <-phpStart.ClickedCh:
				fmt.Println("---- PHP start ---- ")
				phpServer.Start()
				php.Check()
			case <-phpStop.ClickedCh:
				fmt.Println("---- PHP stop ---- ")
				phpServer.Stop()
				php.Uncheck()
			case <-phpRestart.ClickedCh:
				fmt.Println("---- PHP restart ---- ")
				phpServer.Stop()
				phpServer.Start()
				php.Check()
			case <-nginxStart.ClickedCh:
				fmt.Println("---- nginx start ---- ")
				nginxServer.Start()
				nginx.Check()
			case <-nginxStop.ClickedCh:
				fmt.Println("---- nginx stop ---- ")
				nginxServer.Stop()
				nginx.Uncheck()
			case <-nginxRestart.ClickedCh:
				fmt.Println("---- nginx restart ---- ")
				nginxServer.Stop()
				nginxServer.Start()
				nginx.Check()
			}
			setTempIcon()
		}
	}()
}

func initRootPath() string {
	root := os.Getenv("GWNP_ROOT")
	if root == "" {
		// 获取当前执行文件的目录
		root, _ = os.Getwd()
	}
	return root
}

// func onTest() {
// 	startIcon, _ := os.ReadFile("start.ico")
// 	systray.SetIcon(startIcon)
// 	setTempIcon := func() {
// 		if nginxServer.Status() {
// 			// startIcon, _ := os.ReadFile("start.ico")
// 			// systray.SetIcon(startIcon)
// 		}
// 	}
// 	go func() {
// 		setTempIcon()
// 	}()
// 	systray.SetTitle("GNP")
// 	systray.SetTooltip("GNP")
// 	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
// 	go func() {
// 		for {
// 			<-mQuit.ClickedCh
// 			systray.Quit()
// 		}
// 	}()
// }

func readPHPDIR(phpPath string) PhpVersionList {
	phpVersions := make(PhpVersionList, 0)
	file, err := os.ReadDir(phpPath)
	if err != nil {
		fmt.Println(err)
		return phpVersions
	}
	//把servers转为json
	//写入settings.json文件

	for _, f := range file {
		if f.IsDir() {
			if strings.HasPrefix(f.Name(), "php") {
				phpVersions = append(phpVersions, cmd.PhpVersionEntry{Name: f.Name(), Active: false})
			}
		}
	}
	return phpVersions
}

func writeToJson(servers PhpVersionList, configFile string) {
	jsonData, err := json.Marshal(servers)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 写入之前如何格式化json字符串
	// 使用 json.Indent 实现格式化输出
	var prettyJSON []byte
	buffer := new(bytes.Buffer)
	err = json.Indent(buffer, jsonData, "", "    ")
	if err != nil {
		fmt.Println("格式化 JSON 错误:", err)
		return
	}
	prettyJSON = buffer.Bytes()
	jsonData = prettyJSON

	os.WriteFile(configFile, jsonData, 0644)
	fmt.Println(configFile, "写入成功")
}

func readJson(configFile string) PhpVersionList {
	// 读取settings.json文件
	phpVersions := make(PhpVersionList, 0)
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
		return phpVersions
	}
	// fmt.Println(string(data))
	err = json.Unmarshal(data, &phpVersions)
	if err != nil {
		fmt.Println(err)
		return phpVersions
	}
	return phpVersions
}
