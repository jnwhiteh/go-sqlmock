package sqlmock

import (
	"database/sql/driver"
)

type statement struct {
	mockConn *mockConn
	query    string
}

func (stmt *statement) Close() error {
	return nil
}

func (stmt *statement) NumInput() int {
	return -1
}

func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.mockConn.Exec(stmt.query, args)
}

func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.mockConn.Query(stmt.query, args)
}
