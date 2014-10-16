package sqlmock

import "regexp"

// MockDB is returned by the sqlmock package and is used to specify and
// validate expectations.
type MockDB struct {
	expectations []expectation
	active       expectation
}

// ExpectBegin expects transaction to be started
func (m *MockDB) ExpectBegin() *expectedBegin {
	e := &expectedBegin{}
	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}

// ExpectCommit expects transaction to be commited
func (m *MockDB) ExpectCommit() *expectedCommit {
	e := &expectedCommit{}
	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}

// ExpectRollback expects transaction to be rolled back
func (m *MockDB) ExpectRollback() *expectedRollback {
	e := &expectedRollback{}
	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}

// ExpectPrepare expects Query to be prepared
func (m *MockDB) ExpectPrepare() *expectedPrepare {
	e := &expectedPrepare{}
	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}

// ExpectExec expects database Exec to be triggered, which will match
// the given query string as a regular expression
func (m *MockDB) ExpectExec(sqlRegexStr string) *expectedExec {
	e := &expectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}

// ExpectQuery database Query to be triggered, which will match
// the given query string as a regular expression
func (m *MockDB) ExpectQuery(sqlRegexStr string) *expectedQuery {
	e := &expectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)

	m.expectations = append(m.expectations, e)
	m.active = e
	return e
}
