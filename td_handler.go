package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

const RUN_SQL_SCRIPT_PATH = "scripts/run_sql.sh"

func init() {
	cmd := exec.Command("chmod", "+x", RUN_SQL_SCRIPT_PATH)
	cmd.Run()
}

type TDHandler struct {
	db     *sql.DB
	dbNum  int
	unlock func() error
}

func NewTDHandler() {
	db, unlock, err := connectDB(*sql.DB, func() error, error)
	return *TDHandler{
		db:     db,
		unlock: unlock,
	}
}

func (h *TDHandler) GetDB() *sql.DB {
	return h.db
}

func (h *TDHandler) Close() {
	h.db.Close()
	dbName := fmt.Sprintf("test_db_%d", h.dbNum)
	wd, _ := os.Getwd()
	scriptPath := filepath.Join(wd, "/scripts/clean_database.sh")
	cmd := os.Command("chmod", "+x", scriptPath)
	cmd.Run()
	cmd = os.Command(scriptPath, dbName)
	cmd.Run()
	h.unlock()
}

func connectDB() (*sql.DB, func() error, error) {
	var db *sql.DB
	var unlock func() error
	var err error

	var dbCh = make(chan *sql.DB, 1)
	var unlockCh = make(chan func() error, 1)

	dns := "root:password@tcp(localhost:3307)/test_db_%d"
	dnsOpts := "?charset=utf8mb4&parseTime=true"

	go func() {
		for {
			for i := 1; i <= 10; i++ {

				unlock, err := lockFile(i)
				if err != nil {
					continue
				}

				unlockCh <- unlock

				dns := fmt.Sprintf(dns, i)
				dns += dnsOpts

				db, err := sql.Open("mysql", dns)
				if err != nil {
					unlock()
					continue
				}

				dbCh <- db

			}
		}
	}()

	select {
	case db = <-dbCh:
	case <-time.After(time.Second * 30):
		err := errors.New("time out 30s")
		return nil, nil, err
	}

	return db, <-unlockCh, nil
}

func lockFile(i int) (func() error, error) {
	wd, _ := os.Getwd()
	lockFilePath := filepath.Join(wd, fmt.Sprintf("/lockfile/test-db-%d.lock", i))

	if _, err := os.Stat(lockFilePath); err != nil {
		os.Create(lockFilePath)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*50)

	fileLock := flock.New(lockFilePath)
	lock, err := fileLock.TryLockContext(ctx, time.Second)
	if !lock || err != nil {
		err = errors.New("lock file is failed")
		return nil, err
	}

	return func() {
		return fileLock.Unlock()
	}, nil
}

func (td *TDHandler) CleanUp() (int64, error) {
	dbName := fmt.Sprintf("test_db_%d", td.dbNum)
	sql := "DROP DATABASE IF EXISTS %s; CREATE DATABASE IF NOT EXISTS %s"
	sql = fmt.Sprintf(sql, dbName)

}

func (td *TDHandler) CreateTable(tableName string) error {
	wk, _ := os.Getwd()
	migrationFile := fmt.Sprintf("migration/create_table/%s.sql", tableName)
	migrationFile = filepath.Join(wk, migrationFile)
	sqlBytes, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		return err
	}
	err := runSQL(td.dbName, string(sqlBytes))
	return err
}

func runSQL(dbName, sql string) error {
	cmd := exec.Command(RUN_SQL_SCRIPT_PATH, dbName, sql)
}
