package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

// mockConn is an implementation of the database/sql/driver.Conn interface. It
// is designed to be used behind a sql.DB rather than accessed directly.
type mockConn struct {
	expectations []expectation
	active       expectation
}

// next returns the next unfulfilled expectation for this connection
func (c *mockConn) next() (e expectation) {
	for _, e = range c.expectations {
		if !e.fulfilled() {
			return
		}
	}
	return nil // all expectations were fulfilled
}

// Close will close the mock database connection and ensures that all
// expectations were met successfully.
func (c *mockConn) Close() (err error) {
	for _, e := range c.expectations {
		if !e.fulfilled() {
			err = fmt.Errorf("there is a remaining expectation %T which was not matched yet", e)
			break
		}
	}
	c.expectations = nil
	c.active = nil
	return err
}

func (c *mockConn) Begin() (driver.Tx, error) {
	e := c.next()
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to begin transaction was not expected")
	}

	etb, ok := e.(*ExpectedBegin)
	if !ok {
		return nil, fmt.Errorf("call to begin transaction, was not expected, next expectation is %T as %+v", e, e)
	}
	etb.triggered = true
	return &transaction{c}, etb.err
}

func (c *mockConn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to exec '%s' query with args %+v was not expected", query, args)
	}

	eq, ok := e.(*ExpectedExec)
	if !ok {
		return nil, fmt.Errorf("call to exec query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.result == nil {
		return nil, fmt.Errorf("exec query '%s' with args %+v, must return a database/sql/driver.result, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("exec query '%s', does not match regex '%s'", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("exec query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	return eq.result, err
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	e := c.next()

	// for backwards compatibility, ignore when Prepare not expected
	if e == nil {
		return &statement{c, stripQuery(query)}, nil
	}
	eq, ok := e.(*ExpectedPrepare)
	if !ok {
		return &statement{c, stripQuery(query)}, nil
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	return &statement{c, stripQuery(query)}, nil
}

func (c *mockConn) Query(query string, args []driver.Value) (rw driver.Rows, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to query '%s' with args %+v was not expected", query, args)
	}

	eq, ok := e.(*ExpectedQuery)
	if !ok {
		return nil, fmt.Errorf("call to query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.rows, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("query '%s', does not match regex [%s]", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	return eq.rows, err
}

func argMatcherErrorHandler(errp *error) {
	if e := recover(); e != nil {
		if se, ok := e.(*reflect.ValueError); ok { // catch reflect error, failed type conversion
			*errp = fmt.Errorf("Failed to compare query arguments: %s", se)
		} else {
			panic(e) // overwise panic
		}
	}
}
