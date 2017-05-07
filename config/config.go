package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const settingFileName = "config.toml"

//EnvConfig 環境設定構造体
type EnvConfig struct {
	Log    LogEnvConfig
	Server ServerEnvConfig
	DB     DBEnvConfig
	Login  LoginConfig
}

//LogEnvConfig ログ設定情報
type LogEnvConfig struct {
	Output string
}

//ServerEnvConfig HTTPサーバー設定情報
type ServerEnvConfig struct {
	PortNum  int
	HostName string
}

//DBEnvConfig DB接続設定情報
type DBEnvConfig struct {
	UserID     string
	Password   string
	HostName   string
	PortNumber string
	Name       string
}

//LoginConfig ログイン設定情報
type LoginConfig struct {
	PassSalt  string
	TokenSalt string
}

// 設定情報保持変数
var envConfig = EnvConfig{
	Log:    LogEnvConfig{Output: "stream"},
	Server: ServerEnvConfig{PortNum: 8080, HostName: "localhost:8080"},
	DB:     DBEnvConfig{UserID: "root", Password: "root", HostName: "127.0.0.1", PortNumber: "3306", Name: "squidgirl"},
	Login:  LoginConfig{PassSalt: "Cp0xtdDLsHpdadfxysuemBr5a55EDgVv4hzZGyRP", TokenSalt: "Jz2tS4HdzWRNdWbD46SemE6Eh5LZUY2EVGcpkbRx"},
}

//init 初期化
func init() {
}

//GetConfig 現在の設定値を取得する
func GetConfig() EnvConfig {
	return envConfig
}

//LoadConfig 設定ファイルから設定を読み込み（失敗時はfalseを返す）
func LoadConfig() bool {
	var config EnvConfig
	if _, err := toml.DecodeFile(settingFileName, &config); err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Printf("config=%#v\n", config)
	envConfig = config
	return true
}
