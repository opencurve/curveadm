package driver

import (
	"strings"
	"sync"

	"github.com/rqlite/gorqlite"
	rqlite "github.com/rqlite/gorqlite"
)

type RQLiteDB struct {
	conn *rqlite.Connection
	sync.Mutex
}

type QueryResult struct {
	result rqlite.QueryResult
}

type WriteResult struct {
	result rqlite.WriteResult
}

var (
	_ IDataBaseDriver = (*RQLiteDB)(nil)
	_ IQueryResult    = (*QueryResult)(nil)
	_ IWriteResult    = (*WriteResult)(nil)
)

func NewRQLiteDB() *RQLiteDB {
	return &RQLiteDB{}
}

func (db *RQLiteDB) Open(url string) error {
	connURL := "http://" + strings.TrimPrefix(url, "rqlite://")
	conn, err := gorqlite.Open(connURL)
	if err != nil {
		return err
	}
	db.conn = conn
	return nil
}

func (db *RQLiteDB) Close() error {
	return nil
}

func (result *QueryResult) Next() bool {
	return result.result.Next()
}

func (result *QueryResult) Scan(dest ...any) error {
	return result.result.Scan(dest...)
}

func (result *QueryResult) Close() error {
	return nil
}

func (db *RQLiteDB) Query(query string, args ...any) (IQueryResult, error) {
	db.Lock()
	defer db.Unlock()

	result, err := db.conn.QueryOneParameterized(
		rqlite.ParameterizedStatement{
			Query:     query,
			Arguments: append([]interface{}{}, args...),
		},
	)
	return &QueryResult{result: result}, err
}

func (result *WriteResult) LastInsertId() (int64, error) {
	return result.result.LastInsertID, nil
}

func (db *RQLiteDB) Write(query string, args ...any) (IWriteResult, error) {
	db.Lock()
	defer db.Unlock()

	result, err := db.conn.WriteOneParameterized(
		rqlite.ParameterizedStatement{
			Query:     query,
			Arguments: append([]interface{}{}, args...),
		},
	)
	return &WriteResult{result: result}, err
}
