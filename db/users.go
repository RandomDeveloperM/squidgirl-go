package db

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

//UserTable ユーザー情報テーブル
type UserTable struct {
	ID         int64     `db:"id"`
	Name       string    `db:"name"`
	PassHash   string    `db:"passhash"`
	Permission int       `db:"permission"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

//定数
const (
	UserPermissionUser  = 1
	UserPermissionAdmin = 100
)

func InsertUser(name string, password string, permission int) error {
	if name == "" || password == "" || permission == 0 {
		return fmt.Errorf("パラメーターエラー")
	}

	passHash := CreatePasswordHash(password)
	record := UserTable{Name: name, PassHash: passHash, Permission: permission, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err := insertUser(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUser(name string, password string, permission int) error {
	if name == "" || password == "" || permission == 0 {
		return fmt.Errorf("パラメーターエラー")
	}

	passHash := CreatePasswordHash(password)
	record := UserTable{Name: name, PassHash: passHash, Permission: permission, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err := updateUser(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func SelectUser(name string) (UserTable, error) {
	var result UserTable
	recordList, err := selectUserList(nil, name)
	if err != nil {
		return result, err
	}

	if len(recordList) == 0 {
		return result, nil
	}
	return recordList[0], nil
}

func SelectUserAll() ([]UserTable, error) {
	recordList, err := selectUserListAll(nil)
	if err != nil {
		return nil, err
	}
	return recordList, nil
}

func DeleteUser(id int64) error {
	if id == 0 {
		return fmt.Errorf("パラメーターエラー")
	}

	err := deleteUser(nil, id)
	if err != nil {
		return err
	}
	return nil
}

func insertUser(session *dbr.Session, record UserTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.InsertInto(userTableName).
		Columns("name", "passhash", "permission", "created_at", "updated_at").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateUser(session *dbr.Session, record UserTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.Update(userTableName).
		Set("passhash", record.PassHash).
		Set("permission", record.Permission).
		Set("updated_at", record.UpdatedAt).
		Where("name = ?", record.Name).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectUserList(session *dbr.Session, name string) ([]UserTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []UserTable
	_, err = session.Select("*").From(userTableName).Where("name = ?", name).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func selectUserListAll(session *dbr.Session) ([]UserTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []UserTable
	_, err = session.Select("*").From(userTableName).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func deleteUser(session *dbr.Session, id int64) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.DeleteFrom(userTableName).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

//CreatePasswordHash パスワードからパスワードハッシュを生成する
func CreatePasswordHash(password string) string {
	hashBytes := sha256.Sum256([]byte(config.GetConfig().Login.PassSalt + password))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}
