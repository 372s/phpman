package cmd

import (
	"database/sql"
	"fmt"
	"phpman/util"
	"log"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// https://pkg.go.dev/github.com/ncruces/go-sqlite3
// https://pkg.go.dev/github.com/ncruces/go-sqlite3/driver#example-package

type Server struct {
	Name string
	Version string
	Path string
	Port string
	Active int
}

var utilLog = util.NewUtilLog()

func getRootPath() string {
	dir := util.UtilDirectory{}
	return dir.RootPath()
}

func connect() (*sql.DB) {
	wd := getRootPath()
	path := filepath.Join(wd, "settings.db")
	db,err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Println("sql.open error:", err)
		utilLog.Log("sql.open error:" + err.Error())
	}
	return db
}

func CreateSettionsTable() {
	db := connect()
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		port INTEGER NOT NULL,
		path TEXT NOT NULL,
		active INTEGER NOT NULL,
		UNIQUE (name, version)
	)`)
	if err != nil {
		log.Println("Error creating servers table:", err)
		utilLog.Log(fmt.Sprintf("Error creating servers table: %s", err))
	}
	defer db.Close()
}

func GetActiveServersFromDB(name string) []Server {
	servers := make([]Server, 0)
	db := connect()
	rows, err := db.Query(`SELECT name, version, active, path, port FROM servers WHERE name = ? and active = 1`, name)
	if err != nil {
		log.Println("Error querying servers:", err)
		return servers
	}
	defer rows.Close()
	for rows.Next() {
		var active int
		var name, version, path, port string
		err := rows.Scan(&name, &version, &active, &path, &port)
		if err != nil {
			log.Println("Error scanning servers:", err)
			continue
		}
		// log.Println("name:", name, "Version:", version, "Active:", active, "Path:", path, "Port:", port)
		servers = append(servers, Server{Name: name, Version: version, Active: active, Path: path, Port: port})
	}
	defer db.Close()
	return servers
}

func GetServerRow(name string, version string) (Server, error) {
	db := connect()
	row := db.QueryRow(`SELECT id, active, path, port FROM servers WHERE name = ? and version = ?`, name, version)
	if row.Err() != nil {
		fmt.Println("no rows returned", row.Err())
		return Server{}, row.Err()
	}
	var id, active int
	var path, port string
	err := row.Scan(&id, &active, &path, &port)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有找到记录
			fmt.Println("没有找到匹配的记录")
			return Server{}, sql.ErrNoRows
		} else {
			// 其他错误
			fmt.Printf("查询错误: %v\n", err)
			return Server{}, fmt.Errorf("查询错误")
		}
	}
	// log.Println("name:", name, "Version:", version, "Active:", active, "Path:", path, "Port:", port)
	defer db.Close()
	return Server{Name: name, Path: path, Version: version, Port: port, Active: active}, nil
}

func GetActiveServerRow(name string, active int) (Server, error) {
	db := connect()
	row := db.QueryRow(`SELECT path, port, version FROM servers WHERE name = ? and active = ?`, name, active)
	if row.Err() != nil {
		fmt.Println("no rows returned", row.Err())
		return Server{}, row.Err()
	}
	var path, port, version string
	err := row.Scan(&path, &port, &version)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有找到记录
			fmt.Println("没有找到匹配的记录")
			return Server{}, sql.ErrNoRows
		} else {
			// 其他错误
			fmt.Printf("查询错误: %v\n", err)
			return Server{}, fmt.Errorf("查询错误")
		}
	}
	defer db.Close()
	return Server{Name: name, Path: path, Version: version, Port: port, Active: active}, nil
}

func SaveServerToDB(server Server) {
	db := connect()
	query := "INSERT INTO servers (name, version, active, path, port) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, server.Name, server.Version, server.Active, server.Path, server.Port)
	if err != nil {
		log.Println("Insert server error:", err)
	}
	db.Close()
}

func UpdateServerInDB(server Server) { 
	db := connect()
	query := "UPDATE servers SET active = ? WHERE name = ? AND version = ?"
	_, err := db.Exec(query, server.Active, server.Name, server.Version)
	if err != nil {
		log.Println("Update server error:", err)
	}
	db.Close()
}


func UpdateNginxServerActiveInDB(active int, version string) { 
	db := connect()
	query := `
	UPDATE servers SET active = 0 WHERE name = 'nginx';
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Update server error:", err)
	}
	query = `
	UPDATE servers SET active = ? WHERE name = 'nginx' AND version = ?;
	`
	_, err = db.Exec(query, active, version)
	if err != nil {
		log.Println("Update server error:", err)
	}
	db.Close()
}