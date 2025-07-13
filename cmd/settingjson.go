package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PhpVersionEntry struct {
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Path    string `json:"path"`
	Port    string `json:"port"`
	Version string `json:"version"`
}

type PhpVersionList []PhpVersionEntry

func createJson() {
	root:=""
	phpPath:=""
	configFile := filepath.Join(root, "settings.json")
	// 判断是否有settings.json文件
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		phpList := readPhpDIR(phpPath)
		writeToJson(phpList, configFile)
	}
}


func readPhpDIR(phpPath string) PhpVersionList {
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
			// 匹配php目录：php-7.4.25-Win32-vc15-x64
			re := regexp.MustCompile(`^php-\d+\.\d+\.\d+`)
			if strings.HasPrefix(f.Name(), "php") && re.MatchString(f.Name()) {
				// php-8.1.30-nts-Win32-vc15-x64
				filename := f.Name()
				version := filename[4:7]
				// 删掉版本号中的.
				port := "90" + strings.ReplaceAll(version, ".", "")
				phpVersions = append(phpVersions, PhpVersionEntry{Name: "php", Path:f.Name(), Port:port, Version: version, Active: false})
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