//go:build windows
// +build windows

package main

import (
	"database/sql"
	"fmt"
	"phpman/cmd"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getlantern/systray"
)

type Env struct {
	Root              string
	BinPath           string
	NginxPath         string
	PhpPath           string
	PhpCgiSpawnerFile string
	IconPath          string
	PHPServer         *cmd.PHPServer
	NginxServer       *cmd.NginxServer
	PHPServerMenuItem *systray.MenuItem
	NginxServerMenuItem *systray.MenuItem
}

type PhpVersionList []cmd.Server

type NginxList []cmd.Server

type updateNginxMenuItem struct {
	index int
	item  *systray.MenuItem
}

type UpdateSelectedServer struct {
	index int
	Server cmd.Server
	// Action string
}

var updateSelectedServerCh = make(chan UpdateSelectedServer)

func initRootPath() string {
	root := os.Getenv("GWNP_ROOT")
	if root == "" {
		// 获取当前执行文件的目录
		root, _ = os.Getwd()
	}
	return root
}

func main() {
	MustCreateTable()

	systray.Run(onReady, nil)
}

func onReady() {
	// root := "D:\\phpserv"
	// util.NewUtilLog().Debug(fmt.Sprintf("GWNP_ROOT: %v", os.Getenv("GWNP_ROOT")))
	// utilLog.Debug(fmt.Sprintf("Path: %v", os.Getenv("Path")))
	root := initRootPath()
	phpCgiSpawnerFile := filepath.Join(root, "bin", "php-cgi-spawner", "php-cgi-spawner.exe")
	phpPath := filepath.Join(root, "servers", "php")
	nginxPath := filepath.Join(root, "servers", "nginx")
	env := &Env{
		Root:              root,
		BinPath:           filepath.Join(root, "bin"),
		IconPath:          filepath.Join(root, "icon"),
		NginxPath:         nginxPath,
		PhpPath:           phpPath,
		PhpCgiSpawnerFile: phpCgiSpawnerFile,
		PHPServer: &cmd.PHPServer{
			PhpCgiSpawnerFile: phpCgiSpawnerFile, 
			PhpPath: phpPath,
			Versions: getPhpVersions(phpPath),
		},
		NginxServer: &cmd.NginxServer{
			NginxPath: nginxPath,
			Versions: getNginxVersions(nginxPath),
		},
	}

	// 配置菜单
	systray.SetTitle("NP")
	systray.SetTooltip("NP")
	setTempIcon(env)

	go watchServerSelected(env)

	// 创建php版本选择菜单
	systray.AddMenuItem("选择启用的PHP版本(可多选)", "").Disable()
	setPhpSelectMenuItem(env)

	// 创建Nginx版本选择菜单
	systray.AddMenuItem("选择启用的Nginx版本(单选)", "").Disable()
	setNginxSelectMenuItem(env)

	// 添加一个分隔符
	systray.AddSeparator()

	// 添加一个菜单项
	systray.AddMenuItem("启动服务", "启动服务").Disable()

	// 操作php 服务
	setPhpServerMenuItem(env)

	// 操作nginx服务
	setNginxServerMenuItem(env)

	// 操作php+nginx 服务
	// systray.AddMenuItem("启动服务", "启动服务").Disable()
	setAllServersMenuItems(env)

	// 退出
	quit()
}

func setTempIcon(env *Env) {
	if env.PHPServer.Status() && env.NginxServer.Status() {
		startIcon, _ := os.ReadFile(filepath.Join(env.IconPath, "start.ico"))
		systray.SetTemplateIcon(startIcon, startIcon)
	} else if env.PHPServer.Status() || env.NginxServer.Status() {
		someoneIco, _ := os.ReadFile(filepath.Join(env.IconPath, "someone.ico"))
		systray.SetTemplateIcon(someoneIco, someoneIco)
	} else {
		stopIco, _ := os.ReadFile(filepath.Join(env.IconPath, "stop.ico"))
		systray.SetTemplateIcon(stopIco, stopIco)
	}
}

func MustCreateTable() {
	cmd.CreateSettionsTable()
}

func getNginxVersions(dir string) []cmd.Server { 
	nginxList := readNginxDIR(dir)
	// util.NewUtilLog().Log(fmt.Sprintf("nginxList: %v", nginxList))
	return saveServers(nginxList)
}

func getPhpVersions(phpDir string) []cmd.Server { 
	phpList := readPhpDIR(phpDir)
	// util.NewUtilLog().Log(fmt.Sprintf("phpList: %v", phpList))
	return saveServers(phpList)
}


func watchServerSelected(env *Env) {
	phpVersions := env.PHPServer.Versions
	nginxList := &env.NginxServer.Versions
	for req := range updateSelectedServerCh { 
		log.Println("watchServerSelected : ", req.Server)
		// 解引用指针后进行索引操作
		// cmd.UpdateServerInDB(req.Server)
		if req.Server.Name == "nginx" { 
			if req.index >= 0 && req.index < len(*nginxList) {
				(*nginxList)[req.index] = req.Server
			}
			for idx, nginx := range *nginxList {
				if idx != req.index && req.Server.Active == 1 {
					nginx.Active = 0
					(*nginxList)[idx] = nginx
				}
			}
			log.Println("nginx list : ", *nginxList)
			env.NginxServer.Versions = *nginxList
			cmd.UpdateNginxServerActiveInDB(req.Server.Active, req.Server.Version)
		} else { 
			if req.index >= 0 && req.index < len(phpVersions) {
				phpVersions[req.index] = req.Server
			}
			env.PHPServer.Versions = phpVersions
			cmd.UpdateServerInDB(req.Server)
		}
		
	}
}


func setPhpSelectMenuItem(env *Env) {
	phpVersions := env.PHPServer.Versions
	for index, php := range phpVersions {
		tray := systray.AddMenuItemCheckbox(php.Path, php.Path, (php.Active == 1))
		go func(tray *systray.MenuItem, php cmd.Server, idx int) {
			for {
				<-tray.ClickedCh
				if tray.Checked() {
					tray.Uncheck()
					php.Active = 0
				} else {
					tray.Check()
					php.Active = 1
				}
				updateSelectedServerCh <- UpdateSelectedServer{idx, php}
				// fmt.Println(php)
				// cmd.UpdateServerInDB(php)
			}
		}(tray, php, index)
	}
}

func setNginxSelectMenuItem(env *Env) {
	nginxList := env.NginxServer.Versions
	updateNginxChan := make(chan updateNginxMenuItem)
	nginxBoxes := make([]*systray.MenuItem, 0)

	go func() {
		// for req := range updateNginxChan {
		// 	fmt.Println("req: ", req.item)
		// 	nginxBoxes[req.index] = req.item
		// }
		for {
			req, ok := <-updateNginxChan
			if !ok {
				break // channel 已关闭
			}
			log.Println("update nginx menu item req: ", req.item)
			nginxBoxes[req.index] = req.item
		}
	}()

	for _, nginxserver := range nginxList {
		nginxCheckBox := systray.AddMenuItemCheckbox(nginxserver.Path, nginxserver.Path, (nginxserver.Active == 1))
		nginxBoxes = append(nginxBoxes, nginxCheckBox)
	}

	for index, nginxCheckBox := range nginxBoxes {
		nginxserver := nginxList[index]
		go func(tray *systray.MenuItem, nginxserver cmd.Server, idx int) {
			for {
				<-tray.ClickedCh
				// fmt.Println("Clicked on", tray, tray.Checked())
				if tray.Checked() {
					nginxserver.Active = 0
					tray.Uncheck()
				} else {
					nginxserver.Active = 1
					for _, nginxCheckBox := range nginxBoxes {
						if nginxCheckBox != tray {
							nginxCheckBox.Uncheck()
						}
					}
					tray.Check()
				}
				updateNginxChan <- updateNginxMenuItem{idx, tray}
				updateSelectedServerCh <- UpdateSelectedServer{idx, nginxserver}
				// fmt.Println("Clicked on", tray, tray.Checked())
			}
		}(nginxCheckBox, nginxserver, index)
	}
}

func setPhpServerMenuItem(env *Env) {
	// 检查php是否启动h状态
	phpMenuItem := systray.AddMenuItemCheckbox("php", "php", env.PHPServer.Status())
	env.PHPServerMenuItem = phpMenuItem
	phpServerStart := phpMenuItem.AddSubMenuItem("start", "start")
	phpServerStop := phpMenuItem.AddSubMenuItem("stop", "stop")
	phpServerRestart := phpMenuItem.AddSubMenuItem("restart", "restart")
	go func() {
		for {
			select {
			case <-phpServerStart.ClickedCh:
				serverHandle(env, "php", "start")
			case <-phpServerStop.ClickedCh:
				serverHandle(env, "php", "stop")
			case <-phpServerRestart.ClickedCh:
				serverHandle(env, "php", "restart")
			}
			if env.PHPServer.Status() {
				phpMenuItem.Check()
			} else {
				phpMenuItem.Uncheck()
			}
			env.PHPServerMenuItem = phpMenuItem
			setTempIcon(env)
		}
	}()
}

func setNginxServerMenuItem(env *Env) { 
	// 检查nginx是否启动
	env.NginxServerMenuItem = systray.AddMenuItemCheckbox("nginx", "nginx", env.NginxServer.Status())
	nginxServerStart := env.NginxServerMenuItem.AddSubMenuItem("start", "start")
	nginxServerStop := env.NginxServerMenuItem.AddSubMenuItem("stop", "stop")
	nginxServerRestart := env.NginxServerMenuItem.AddSubMenuItem("restart", "restart")
	go func() {
		for {
			select {
			case <-nginxServerStart.ClickedCh:
				serverHandle(env, "nginx", "start")
				env.NginxServerMenuItem.Check()
			case <-nginxServerStop.ClickedCh:
				serverHandle(env, "nginx", "stop")
				env.NginxServerMenuItem.Uncheck()
			case <-nginxServerRestart.ClickedCh:
				serverHandle(env, "nginx", "restart")
				env.NginxServerMenuItem.Check()
			}
			setTempIcon(env)
		}
	}()
}

func setAllServersMenuItems(env *Env) {
	all := systray.AddMenuItem("所有服务", "")
	allStart := all.AddSubMenuItem("start", "start")
	allStop := all.AddSubMenuItem("stop", "stop")
	allRestart := all.AddSubMenuItem("restart", "restart")
	go func() {
		for {
			select {
			case <-allStart.ClickedCh:
				serverHandle(env, "", "start")
				env.PHPServerMenuItem.Check()
				env.NginxServerMenuItem.Check()
			case <-allStop.ClickedCh:
				serverHandle(env, "", "stop")
				env.PHPServerMenuItem.Uncheck()
				env.NginxServerMenuItem.Uncheck()
			case <-allRestart.ClickedCh:
				serverHandle(env, "", "restart")
				env.PHPServerMenuItem.Check()
				env.NginxServerMenuItem.Check()
			}

			// select {
			// case <-env.PHPServerMenuItem:
			// 	php.Check()
			// case <-env.PHPServerMenuItem:
			// 	php.Uncheck()
			// case <-env.PHPServerMenuItem:
			// 	nginx.Check()
			// case <-env.PHPServerMenuItem:
			// 	nginx.Uncheck()
			// default:
			// }

			setTempIcon(env)
		}
	}()
}

func serverHandle(env *Env, name, action string) {
	log.Println("phpserver: ", env.PHPServer)
	log.Println("nginxserver: ", env.NginxServer)
	switch action {
	case "start":
		switch name {
		case "php":
			// env.PHPServer.Versions = cmd.GetActiveServersFromDB("php")
			env.PHPServer.Start()
		case "nginx":
			// iServer, _ := cmd.GetActiveServerRow("nginx", 1)
			// env.NginxServer.VersionPath = iServer.Path
			env.NginxServer.Start()
		default:
			env.PHPServer.Start()
			// iServer, _ := cmd.GetActiveServerRow("nginx", 1)
			// env.NginxServer.VersionPath = iServer.Path
			env.NginxServer.Start()
		}
	case "stop":
		switch name {
		case "php":
			env.PHPServer.Stop()
		case "nginx":
			env.NginxServer.Stop()
		default:
			env.PHPServer.Stop()
			env.NginxServer.Stop()
		}
	case "restart":
		switch name {
		case "php":
			env.PHPServer.Stop()
			// env.PHPServer.Versions = cmd.GetActiveServersFromDB("php")
			env.PHPServer.Start()
		case "nginx":
			env.NginxServer.Stop()
			// iServer, _ := cmd.GetActiveServerRow("nginx", 1)
			// env.NginxServer.VersionPath = iServer.Path
			env.NginxServer.Start()
		default:
			env.PHPServer.Stop()
			env.NginxServer.Stop()

			// env.PHPServer.Versions = cmd.GetActiveServersFromDB("php")
			env.PHPServer.Start()

			// iServer, _ := cmd.GetActiveServerRow("nginx", 1)
			// env.NginxServer.VersionPath = iServer.Path
			env.NginxServer.Start()
		}
	}
}

func readPhpDIR(phpPath string) PhpVersionList {
	phpVersions := make(PhpVersionList, 0)
	file, err := os.ReadDir(phpPath)
	if err != nil {
		fmt.Println(err)
		return phpVersions
	}
	for _, f := range file {
		if f.IsDir() {
			// 匹配php目录：php-7.4.25-Win32-vc15-x64
			re := regexp.MustCompile(`^php-\d+\.\d+\.\d+`)
			if strings.HasPrefix(f.Name(), "php") && re.MatchString(f.Name()) {
				// php-8.1.30-nts-Win32-vc15-x64
				filename := f.Name()
				version := filename[4:7]
				// 删掉版本号中的.
				port := "90" + strings.ReplaceAll(version, ".", "")
				phpVersions = append(phpVersions, cmd.Server{Name: "php", Path: f.Name(), Port: port, Version: version, Active: 0})
			}
		}
	}
	return phpVersions
}

func readNginxDIR(path string) NginxList {
	list := make(NginxList, 0)
	file, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return list
	}
	for _, f := range file {
		if f.IsDir() {
			// 匹配php目录：php-7.4.25-Win32-vc15-x64
			re := regexp.MustCompile(`^nginx-\d+\.\d+\.\d+`)
			if re.MatchString(f.Name()) {
				// nginx-1.21.1
				filename := f.Name()
				// 删掉版本号中的.
				version := strings.ReplaceAll(filename, "nginx-", "")
				list = append(list, cmd.Server{Name: "nginx", Path: f.Name(), Version: version, Active: 0})
			}
		}
	}
	return list
}

func saveServers(servers []cmd.Server) []cmd.Server {
	var data []cmd.Server
	for _, server := range servers {
		newserver, err := cmd.GetServerRow(server.Name, server.Version)
		if err == sql.ErrNoRows {
			newserver.Active = server.Active
			newserver.Name = server.Name
			newserver.Path = server.Path
			newserver.Port = server.Port
			newserver.Version = server.Version
			cmd.SaveServerToDB(newserver)
		}
		data = append(data, newserver)
	}
	return data
}


func quit() {
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			<-mQuit.ClickedCh
			systray.Quit()
			return
		}
	}()
}