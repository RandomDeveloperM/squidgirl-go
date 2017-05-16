package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/robfig/cron"

	"github.com/mryp/squidgirl-go/config"
)

func main() {
	//環境設定読み込み
	if !config.LoadConfig() {
		log.Println("設定ファイル読み込み失敗（デフォルト値動作）")
	}
	startCrontab()
	startEchoServer()
}

func startCrontab() {
	interval := config.GetConfig().File.WatchInterval
	fileWatcher = NewFileWatcher()
	if interval > 0 {
		c := cron.New()
		format := fmt.Sprintf("0 */%d * * * *", config.GetConfig().File.WatchInterval)
		c.AddFunc(format, func() {
			fileWatcher.RegistFile()
			fileWatcher.ClearFile()
		})
		c.Start()
	} else {
		fileWatcher.RegistFile() //最初だけ呼び出す
		fileWatcher.ClearFile()
	}
}

func startEchoServer() {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS()) //CORS対応（他ドメインからAJAX通信可能にする）
	if config.GetConfig().Log.Output == "stream" {
	}
	switch config.GetConfig().Log.Output {
	case "stream":
		e.Use(middleware.Logger())
	case "file":
		//未実装
	}

	//ルーティング
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "squidgirl-go")
	})
	e.POST("/login", LoginHandler)
	e.POST("/createuser", CreateUserHandler)

	apiGroup := e.Group("/api")
	apiGroup.Use(middleware.JWT([]byte(config.GetConfig().Login.TokenSalt)))
	apiGroup.POST("/filelist", FileListHandler)
	apiGroup.GET("/thumbnail/:hash", ThumbnailHandler)
	apiGroup.GET("/page/:hash", PageHandler)
	apiGroup.POST("/savebook", SaveBookHandler)
	//apiGroup.POST("/logintest", LoginTestHandler)

	//開始
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(config.GetConfig().Server.PortNum)))
}
