/*
Package sqlmock provides sql driver mock connecection, which allows to test database,
create expectations and ensure the correct execution flow of any database operations.
It hooks into Go standard library's database/sql package.

The package provides convenient methods to mock database queries, transactions and
expect the right execution flow, compare query arguments or even return error instead
to simulate failures. See the example bellow, which illustrates how convenient it is
to work with:


    package main

    import (
        "database/sql"
        "github.com/DATA-DOG/go-sqlmock"
        "testing"
        "fmt"
    )

    // will test that order with a different status, cannot be cancelled
    func TestShouldNotCancelOrderWithNonPendingStatus(t *testing.T) {
		// Open new mock database
		mock, db, err := sqlmock.New()
		if err != nil {
			t.Error("error creating mock")
			return
		}

		// columns to be used for result
		columns := []string{"id", "status"}
		// expect transaction begin
		mock.ExpectBegin()
		// expect query to fetch order, match it with regexp
		mock.ExpectQuery("SELECT (.+) FROM orders (.+) FOR UPDATE").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1"))
		// expect transaction rollback, since order status is "cancelled"
		mock.ExpectRollback()

		// run the cancel order function
		someOrderId := 1
		// call a function which executes expected database operations
		err = cancelOrder(db, someOrderId)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		// ensure all expectations have been met
		mock.CloseTest(t)
	}

*/
package sqlmock

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
)

// A global instance of the database driver
var dbDriver *mockDriver

func init() {
	dbDriver = &mockDriver{
		conns: make(map[string]*mockConn), // a map from DSN to connection
		mu:    new(sync.RWMutex),          // mutex for connMap
	}
	sql.Register("mock", dbDriver)
}

// mockDriver satisfies the sql/driver.Driver interface and is used to
// instantiate new connections to the mock database.
type mockDriver struct {
	conns map[string]*mockConn
	mu    *sync.RWMutex
}

// Open a new connection to the mock database. This connection is not safe for
// use by multiple goroutines at a time, but multiple connections to different
// DSNs are supported.
func (d *mockDriver) Open(dsn string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	c, ok := d.conns[dsn]
	if ok {
		return c, nil
	}

	c = &mockConn{}
	d.conns[dsn] = c
	return c, nil
}

// generateDSN creates a new random string that can be used as a DSN in the
// call to Open.
func generateDSN() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("mocksql://%x", b), nil
}

// Create a new MockDB that can be used to state and verify expectations for
// interaction with a database. For completeness, the Check() method should be
// called on the mock object to validate any outstanding expectations.
func New() (*MockDB, *sql.DB, error) {
	dsn, err := generateDSN()
	if err != nil {
		return nil, nil, err
	}

	// Use the database/sql package to open the new connection
	db, err := sql.Open("mock", dsn)
	if err != nil {
		return nil, nil, err
	}

	// Ping the database to ensure the connection is properly opened
	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	// Grab the underlying DB connection from the driver
	db, err = sql.Open("mock", dsn)
	if err != nil {
		return nil, nil, err
	}
	// Fetch the underlying connection for this DSN
	dbDriver.mu.RLock()
	defer dbDriver.mu.RUnlock()

	mockConn, ok := dbDriver.conns[dsn]
	if !ok {
		return nil, nil, errors.New("Failed when looking up connection")
	}

	mockDB := &MockDB{mockConn}
	return mockDB, db, err
}
