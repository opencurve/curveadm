/*
 *  Copyright (c) 2023 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the Licensele().
 */

/*
 * Project: CurveAdm
 * Created Date: 2023-05-24
 * Author: Jingli Chen (Wine93)
 */

package driver

import (
	"database/sql"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
	sync.Mutex
}

type Rows struct {
	rows *sql.Rows
}

type Result struct {
	result sql.Result
}

var (
	_ IDataBaseDriver = (*SQLiteDB)(nil)
	_ IQueryResult    = (*Rows)(nil)
	_ IWriteResult    = (*Result)(nil)
)

func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{}
}

func (db *SQLiteDB) Open(url string) error {
	var err error
	dataSourceName := strings.TrimPrefix(url, "sqlite://")
	db.db, err = sql.Open("sqlite3", dataSourceName)
	return err
}

func (db *SQLiteDB) Close() error {
	return db.db.Close()
}

func (result *Rows) Next() bool {
	return result.rows.Next()
}

func (result *Rows) Scan(dest ...any) error {
	return result.rows.Scan(dest...)
}

func (result *Rows) Close() error {
	return result.rows.Close()
}

func (db *SQLiteDB) Query(query string, args ...any) (IQueryResult, error) {
	db.Lock()
	defer db.Unlock()

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

func (result *Result) LastInsertId() (int64, error) {
	return result.result.LastInsertId()
}

func (db *SQLiteDB) Write(query string, args ...any) (IWriteResult, error) {
	db.Lock()
	defer db.Unlock()

	stmt, err := db.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec(args...)
	return &Result{result: result}, err
}
