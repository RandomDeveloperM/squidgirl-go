package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocraft/dbr"
	"github.com/mryp/squidgirl-go/config"
)

//テーブル名
const userTableName = "users"

//UsersTable ユーザー情報テーブル
type UsersTable struct {
	ID         int64     `db:"id"`
	Name       string    `db:"name"`
	PassHash   string    `db:"passhash"`
	Permission int       `db:"permission"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

const (
	UserPermissionUser  = 1
	UserPermissionAdmin = 100
)

//InsertUser ユーザーを追加する
func InsertUser(name string, password string, permission int) error {
	if name == "" || password == "" || permission == 0 {
		return fmt.Errorf("パラメーターエラー")
	}
	passHash := CreatePasswordHash(password)
	fmt.Printf("CreatePasswordHash " + passHash + "\n")
	record := UsersTable{Name: name, PassHash: passHash, Permission: permission, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	session := ConnectDB()
	if session == nil {
		return fmt.Errorf("DB接続失敗")
	}
	defer session.Close()

	selectRecords := selectUserRecord(session, name)
	if selectRecords == nil || len(selectRecords) == 0 {
		_, err := session.InsertInto(userTableName).
			Columns("name", "passhash", "permission", "created_at", "updated_at").
			Record(record).
			Exec()
		fmt.Printf("InsertInto err=%v\n", err)
		if err != nil {
			return err
		}
	} else {
		_, err := session.Update(userTableName).
			Set("passhash", passHash).
			Set("permission", permission).
			Where("id = ?", selectRecords[0].ID).
			Exec()
		fmt.Printf("Update err=%v\n", err)
		if err != nil {
			return err
		}
	}

	return nil
}

//selectUserRecord 指定したユーザーのレコードを取得する
func selectUserRecord(session *dbr.Session, userName string) []UsersTable {
	var recordList []UsersTable
	_, err := session.Select("*").From(userTableName).Where("name = ?", userName).Load(&recordList)
	if err != nil {
		fmt.Printf("selectRecord err=" + err.Error())
		return nil
	}

	return recordList
}

//SelectUser 指定したユーザーを取得する
func SelectUser(name string) (result UsersTable) {
	session := ConnectDB()
	if session == nil {
		fmt.Printf("SelectUser DB接続失敗\n")
		return
	}
	defer session.Close()

	table := selectUserRecord(session, name)
	if table == nil {
		return
	}

	result = table[0]
	return
}

//CreatePasswordHash パスワードからパスワードハッシュを生成する
func CreatePasswordHash(password string) string {
	hashBytes := sha256.Sum256([]byte(config.GetConfig().Login.PassSalt + password))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}
