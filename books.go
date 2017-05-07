package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocraft/dbr"
)

//テーブル名
const bookTableName = "books"

//BooksTable アーカイブ情報テーブル
type BooksTable struct {
	ID        int64     `db:"id"`
	Hash      string    `db:"hash"`
	FilePath  string    `db:"file_path"`
	FileSize  int       `db:"file_size"`
	Page      int       `db:"page"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

//InsertBook アーカイブを追加する
func InsertBook(filePath string, fileSize int, page int) error {
	if filePath == "" || fileSize == 0 || page == 0 {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createBookHash(filePath)
	record := BooksTable{Hash: hash, FilePath: filePath, FileSize: fileSize, Page: page, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	session := ConnectDB()
	if session == nil {
		return fmt.Errorf("DB接続失敗")
	}
	defer session.Close()

	_, err := session.InsertInto(bookTableName).
		Columns("hash", "file_path", "file_size", "page", "created_at", "updated_at").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func selectBookRecord(session *dbr.Session, hash string) []BooksTable {
	var recordList []BooksTable
	_, err := session.Select("*").From(bookTableName).Where("hash = ?", hash).Load(&recordList)
	if err != nil {
		fmt.Printf("selectBookRecord err=" + err.Error())
		return nil
	}

	return recordList
}

func SelectBook(filePath string) (result BooksTable) {
	session := ConnectDB()
	if session == nil {
		fmt.Printf("DB接続失敗\n")
		return
	}
	defer session.Close()

	hash := createBookHash(filePath)
	table := selectBookRecord(session, hash)
	if table == nil {
		return
	}

	result = table[0]
	return
}

func createBookHash(filePath string) string {
	hashBytes := sha256.Sum256([]byte(filePath))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}
