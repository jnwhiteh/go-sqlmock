package sqlmock

import "regexp"

// MockDB is returned by the sqlmock package and is used to specify and
// validate expectations.
type MockDB struct {
	c *conn
}

func (m *MockDB) Close() error {
	return m.c.Close()
}

// ExpectBegin expects transaction to be started
func (m *MockDB) ExpectBegin() *ExpectedBegin {
	e := &ExpectedBegin{}
	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}

// ExpectCommit expects transaction to be commited
func (m *MockDB) ExpectCommit() *ExpectedCommit {
	e := &ExpectedCommit{}
	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}

// ExpectRollback expects transaction to be rolled back
func (m *MockDB) ExpectRollback() *ExpectedRollback {
	e := &ExpectedRollback{}
	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}

// ExpectPrepare expects Query to be prepared
func (m *MockDB) ExpectPrepare() *ExpectedPrepare {
	e := &ExpectedPrepare{}
	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}

// ExpectExec expects database Exec to be triggered, which will match
// the given query string as a regular expression
func (m *MockDB) ExpectExec(sqlRegexStr string) *ExpectedExec {
	e := &ExpectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}

// ExpectQuery database Query to be triggered, which will match
// the given query string as a regular expression
func (m *MockDB) ExpectQuery(sqlRegexStr string) *ExpectedQuery {
	e := &ExpectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)

	m.c.expectations = append(m.c.expectations, e)
	m.c.active = e
	return e
}
