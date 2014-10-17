package sqlmock

import (
	"fmt"
)

type transaction struct {
	mockConn *mockConn
}

func (tx *transaction) Commit() error {
	e := tx.mockConn.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to commit transaction was not expected")
	}

	etc, ok := e.(*ExpectedCommit)
	if !ok {
		return fmt.Errorf("call to commit transaction, was not expected, next expectation was %v", e)
	}
	etc.triggered = true
	return etc.err
}

func (tx *transaction) Rollback() error {
	e := tx.mockConn.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to rollback transaction was not expected")
	}

	etr, ok := e.(*ExpectedRollback)
	if !ok {
		return fmt.Errorf("call to rollback transaction, was not expected, next expectation was %v", e)
	}
	etr.triggered = true
	return etr.err
}
