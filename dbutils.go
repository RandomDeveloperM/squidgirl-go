package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" //dbrで使用する
	"github.com/gocraft/dbr"

	"github.com/mryp/squidgirl-go/config"
)

//ConnectDB DB接続
func ConnectDB() *dbr.Session {
	dbConfig := config.GetConfig().DB
	db, err := dbr.Open("mysql", dbConfig.UserID+":"+dbConfig.Password+"@tcp("+dbConfig.HostName+":"+dbConfig.PortNumber+")/"+dbConfig.Name+"?parseTime=true", nil)
	if err != nil {
		fmt.Printf("connectDB err=%v\n", err)
		return nil
	}

	dbsession := db.NewSession(nil)
	return dbsession
}
