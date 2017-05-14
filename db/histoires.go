package db

import (
	"fmt"
	"time"

	"github.com/gocraft/dbr"
)

//テーブル名
const historyTableName = "histoires"

//情報テーブル
type HistoryTable struct {
	ID       int64     `db:"id"`
	UserName string    `db:"user_name"`
	BookHash string    `db:"book_hash"`
	ReadPos  int       `db:"read_pos"`
	Reaction int       `db:"reaction"`
	ModTime  time.Time `db:"mod_time"`
}

func InsertHistory(userName string, bookHash string, readPos int, reaction int) error {
	fmt.Printf("InsertHistory userName=%s, bookHash=%s, readPos=%d, reaction=%d\n", userName, bookHash, readPos, reaction)
	if userName == "" || bookHash == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	record := HistoryTable{UserName: userName, BookHash: bookHash, ReadPos: readPos, Reaction: reaction, ModTime: time.Now()}
	err := insertHistory(nil, record)
	if err != nil {
		fmt.Printf("insertHistory err=%s\n", err)
		return err
	}
	return nil
}

func UpdateHistory(userName string, bookHash string, readPos int, reaction int) error {
	fmt.Printf("UpdateHistory userName=%s, bookHash=%s, readPos=%d, reaction=%d\n", userName, bookHash, readPos, reaction)
	if userName == "" || bookHash == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	record := HistoryTable{UserName: userName, BookHash: bookHash, ReadPos: readPos, Reaction: reaction, ModTime: time.Now()}
	err := updateHistory(nil, record)
	if err != nil {
		fmt.Printf("UpdateHistory err=%s\n", err)
		return err
	}
	return nil
}

func SelectHistory(userName string, bookHash string) (HistoryTable, error) {
	fmt.Printf("SelectHistory userName=%s, bookHash=%s\n", userName, bookHash)
	var result HistoryTable
	recordList, err := selectHistoryList(nil, userName, bookHash)
	if err != nil {
		fmt.Printf("SelectHistory err=%s\n", err)
		return result, err
	}

	if len(recordList) == 0 {
		fmt.Printf("SelectHistory len==0\n")
		return result, nil
	}
	return recordList[0], nil
}

func insertHistory(session *dbr.Session, record HistoryTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil
	}
	defer session.Close()

	_, err = session.InsertInto(historyTableName).
		Columns("user_name", "book_hash", "read_pos", "reaction", "mod_time").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateHistory(session *dbr.Session, record HistoryTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil
	}
	defer session.Close()

	builder := session.Update(historyTableName)
	if record.ReadPos != -1 {
		builder = builder.Set("read_pos", record.ReadPos)
	}
	if record.Reaction != -1 {
		builder = builder.Set("reaction", record.Reaction)
	}
	_, err = builder.Set("mod_time", record.ModTime).
		Where("user_name = ? AND book_hash = ?", record.UserName, record.BookHash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectHistoryList(session *dbr.Session, userName string, bookHash string) ([]HistoryTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []HistoryTable
	_, err = session.Select("*").
		From(historyTableName).
		Where("user_name = ? AND book_hash = ?", userName, bookHash).
		Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}
