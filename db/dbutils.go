package db

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" //dbrで使用する
	"github.com/gocraft/dbr"

	"github.com/mryp/squidgirl-go/config"
)

//ConnectDB DB接続
func ConnectDB() (*dbr.Session, error) {
	dbConfig := config.GetConfig().DB
	db, err := dbr.Open("mysql", dbConfig.UserID+":"+dbConfig.Password+"@tcp("+dbConfig.HostName+":"+dbConfig.PortNumber+")/"+dbConfig.Name+"?parseTime=true", nil)
	if err != nil {
		fmt.Printf("connectDB err=%v\n", err)
		return nil, err
	}

	dbsession := db.NewSession(nil)
	return dbsession, nil
}

func ConnectDBRecheck(session *dbr.Session) (*dbr.Session, error) {
	if session == nil {
		newSession, err := ConnectDB()
		if err != nil {
			return nil, err
		}
		session = newSession
	}

	return session, nil
}
