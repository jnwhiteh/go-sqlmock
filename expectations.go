package sqlmock

import (
	"database/sql/driver"
	"reflect"
	"regexp"
)

// commonExpectation is a set of attributes that are common to all
// expectations.
type commonExpectation struct {
	triggered bool  //whether or not the expectation was triggered
	err       error // an error to be returned when triggered
}

// fulfilled returns whether or not the expectation was fulfilled
func (e *commonExpectation) fulfilled() bool {
	return e.triggered
}

// A ExpectedBegin is triggered when the user calls Begin() on a database
type ExpectedBegin struct {
	commonExpectation
}

// WillReturnError arranges for the triggered expectation to return an error
// result
func (e *ExpectedBegin) WillReturnError(err error) *ExpectedBegin {
	e.err = err
	return e
}

// A RollbackException is triggered when the user calls Rollback() on a
// transaction
type ExpectedRollback struct {
	commonExpectation
}

// WillReturnError arranges for the triggered expectation to return an error
// result
func (e *ExpectedRollback) WillReturnError(err error) *ExpectedRollback {
	e.err = err
	return e
}

// A ExpectedCommit is triggered when the user calls Commit() on a
// transaction
type ExpectedCommit struct {
	commonExpectation
}

// WillReturnError arranges for the triggered expectation to return an error
// result
func (e *ExpectedCommit) WillReturnError(err error) *ExpectedCommit {
	e.err = err
	return e
}

// A PrepareExepectation is triggered by an explicit call to Prepare() a
// statement for the database
type ExpectedPrepare struct {
	commonExpectation
}

func (e *ExpectedPrepare) WillReturnError(err error) *ExpectedPrepare {
	e.err = err
	return e
}

// A argExpectation contains fields and implementations that are common
// to expectations that can take parameters, such as Query() and Exec()
type argExpectation struct {
	sqlRegex *regexp.Regexp // a regular expression to match the query
	args     []driver.Value // the arguments that were passed as parameters
}

// argMatches tests whether or not a list of arguments matches those that are
// expected
func (e *argExpectation) argsMatches(args []driver.Value) bool {
	if nil == e.args {
		return true
	}
	if len(args) != len(e.args) {
		return false
	}
	for k, v := range args {
		vi := reflect.ValueOf(v)
		ai := reflect.ValueOf(e.args[k])
		switch vi.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if vi.Int() != ai.Int() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if vi.Float() != ai.Float() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if vi.Uint() != ai.Uint() {
				return false
			}
		case reflect.String:
			if vi.String() != ai.String() {
				return false
			}
		default:
			// compare types like time.Time based on type only
			if vi.Kind() != ai.Kind() {
				return false
			}
		}
	}
	return true
}

func (e *argExpectation) queryMatches(sql string) bool {
	return e.sqlRegex.MatchString(sql)
}

// A ExpectedQuery is triggered by a call to Query() either directly on the
// database or within a transaction
type ExpectedQuery struct {
	commonExpectation
	argExpectation
	rows driver.Rows // the rows to be returned by this query
}

// WillReturnError arranges for the triggered expectation to return an error
// result
func (e *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	e.err = err
	return e
}

// WithArgs specifies the arguments that are expected when the query is made
func (e *ExpectedQuery) WithArgs(args ...driver.Value) *ExpectedQuery {
	e.args = args
	return e
}

// WillReturnRows specifies the set of resulting rows that will be returned
// by the triggered query
func (e *ExpectedQuery) WillReturnRows(rows driver.Rows) *ExpectedQuery {
	e.rows = rows
	return e
}

// A ExpectedExec is triggered by a call to Exec() either directly on the
// database or within a transaction
type ExpectedExec struct {
	commonExpectation
	argExpectation
	result driver.Result // the result to be returned

}

// WillReturnError arranges for the triggered expectation to return an error
// result
func (e *ExpectedExec) WillReturnError(err error) *ExpectedExec {
	e.err = err
	return e
}

// WithArgs specifies the arguments that are expected when the query is made
func (e *ExpectedExec) WithArgs(args ...driver.Value) *ExpectedExec {
	e.args = args
	return e
}

// WillReturnResult arranges for an expected Exec() to return a particular
// result
func (e *ExpectedExec) WillReturnResult(result driver.Result) *ExpectedExec {
	e.result = result
	return e
}
