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
        // open database stub
        db, err := sql.Open("mock", "")
        if err != nil {
            t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
        }

        // columns to be used for result
        columns := []string{"id", "status"}
        // expect transaction begin
        sqlmock.ExpectBegin()
        // expect query to fetch order, match it with regexp
        sqlmock.ExpectQuery("SELECT (.+) FROM orders (.+) FOR UPDATE").
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1"))
        // expect transaction rollback, since order status is "cancelled"
        sqlmock.ExpectRollback()

        // run the cancel order function
        someOrderId := 1
        // call a function which executes expected database operations
        err = cancelOrder(someOrderId, db)
        if err != nil {
            t.Errorf("Expected no error, but got %s instead", err)
        }
        // db.Close() ensures that all expectations have been met
        if err = db.Close(); err != nil {
            t.Errorf("Error '%s' was not expected while closing the database", err)
        }
    }



func TestMigrationApplied(t *testing.T) {
    // Create a new DB instance and arrange for expectation checks (on defer)
    dbMock, db, err := mocksql.New()
    defer dbMock.Check(t)

    // Arrange for the right response to be returned
    dbMock.ExpectQuery("SELECT version FROM migrate LIMIT 1").
    	WillReturnRows(sqlmock.Rows(columns).FromCSVString("1,1"))

    // Call function that requires the database
    err := emigrate.Migrate(db, 1)
    if err != nil {
    	t.Fatalf("Unexpected error: %s", err)
    }
}

*/

package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
)

// The global instance of the database driver
var dbDriver *mockDriver

type mockDriver struct {
	driverName  string
	sequence    int
	connections map[string]*conn
	mu          *sync.Mutex
}

func init() {
	driverName := fmt.Sprintf("mocksql-%p", new(int))

	// Create a new driver and register with the database/sql package
	dbDriver = &mockDriver{
		driverName:  driverName,
		sequence:    0,
		connections: make(map[string]*conn),
		mu:          new(sync.Mutex),
	}
	sql.Register(driverName, dbDriver)
}

// Open satisfiies the database/sql/driver.Open interface, but is only called
// internally from the New() method. This is accomplished by uniquely
// generating and not publishing the name of the database driver.
func (d *mockDriver) Open(dsn string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	c, ok := d.connections[dsn]
	if ok {
		// This should never happen, but just in case of insanity
		return nil, errors.New("DSN collision error")
	}
	c = &conn{}
	d.connections[dsn] = c
	return c, nil
}

// reserveDSN increments the sequence counter and returns a DSN string that
// will be unique.
func (d *mockDriver) reserveDSN() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	dsn := fmt.Sprintf("mocksql://db/%d", d.sequence)
	d.sequence++
	return dsn
}

// getConnection fetches the connection object that is tied to the given DSN
func (d *mockDriver) getConnection(dsn string) *conn {
	d.mu.Lock()
	defer d.mu.Unlock()

	conn, ok := d.connections[dsn]
	if ok {
		return conn
	}
	return nil
}

// Create a new MockDB that can be used to state and verify expectations for
// interaction with a database.
func New() (*MockDB, *sql.DB, error) {
	// Reserve a DSN for this connection
	dsn := dbDriver.reserveDSN()

	// Use the database/sql package to open the new connection
	db, err := sql.Open(dbDriver.driverName, dsn)
	if err != nil {
		return nil, nil, err
	}

	// Ping the database to ensure the connection is properly opened
	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	// Grab a DB from the sql package
	db, err = sql.Open("mock", dsn)
	if err != nil {
		return nil, nil, err
	}
	// Fetch the underlying connection for this DSN
	conn := dbDriver.getConnection(dsn)
	if conn == nil {
		return nil, nil, errors.New("Failed when looking up connection")
	}

	// TODO: Fix this
	return nil, db, err
}
