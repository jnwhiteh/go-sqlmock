package sqlmock

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func cancelOrder(db *sql.DB, orderId int) error {
	tx, _ := db.Begin()
	_, _ = tx.Query("SELECT * FROM orders {0} FOR UPDATE", orderId)
	_ = tx.Rollback()
	return nil
}

func Example() {
	// Open new mock database
	mock, db, err := New()
	if err != nil {
		fmt.Println("error creating mock")
		return
	}
	// columns to be used for result
	columns := []string{"id", "status"}
	// expect transaction begin
	mock.ExpectBegin()
	// expect query to fetch order, match it with regexp
	mock.ExpectQuery("SELECT (.+) FROM orders (.+) FOR UPDATE").
		WithArgs(1).
		WillReturnRows(NewRows(columns).FromCSVString("1,1"))
	// expect transaction rollback, since order status is "cancelled"
	mock.ExpectRollback()

	// run the cancel order function
	someOrderId := 1
	// call a function which executes expected database operations
	err = cancelOrder(db, someOrderId)
	if err != nil {
		fmt.Printf("unexpected error: %s", err)
		return
	}

	// ensure all expectations have been met
	if err = mock.Close(); err != nil {
		fmt.Printf("unexpected error on close: %s", err)
	}
	// Output:
}

// test the case when db is not triggered and expectations
// are not asserted on close
func TestIssue4(t *testing.T) {
	mock, _, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectQuery("some sql query which will not be called").
		WillReturnRows(NewRows([]string{"id"}))

	err = mock.Close()
	if err == nil {
		t.Errorf("Was expecting an error, since expected query was not matched")
	}
}

func TestMockQuery(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM articles WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Error("it must have had one row as result, but got empty result set instead")
	}

	var id int
	var title string

	err = rows.Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if title != "hello world" {
		t.Errorf("expected mocked title to be 'hello world', but got '%s' instead", title)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestMockQueryTypes(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	columns := []string{"id", "timestamp", "sold"}

	timestamp := time.Now()
	rs := NewRows(columns)
	rs.AddRow(5, timestamp, true)

	mock.ExpectQuery("SELECT (.+) FROM sales WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM sales WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Error("it must have had one row as result, but got empty result set instead")
	}

	var id int
	var time time.Time
	var sold bool

	err = rows.Scan(&id, &time, &sold)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if time != timestamp {
		t.Errorf("expected mocked time to be %s, but got '%s' instead", timestamp, time)
	}

	if sold != true {
		t.Errorf("expected mocked boolean to be true, but got %v instead", sold)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestTransactionExpectations(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// begin and commit
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("an error '%s' was not expected when commiting a transaction", err)
	}

	// begin and rollback
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err = db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("an error '%s' was not expected when rolling back a transaction", err)
	}

	// begin with an error
	mock.ExpectBegin().WillReturnError(fmt.Errorf("some err"))

	tx, err = db.Begin()
	if err == nil {
		t.Error("an error was expected when beginning a transaction, but got none")
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestPrepareExpectations(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// no expectations, w/o mock.ExpectPrepare()
	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}
	if stmt == nil {
		t.Errorf("stmt was expected while creating a prepared statement")
	}

	// expect something else, w/o mock.ExpectPrepare()
	var id int
	var title string
	rs := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	stmt, err = db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}
	if stmt == nil {
		t.Errorf("stmt was expected while creating a prepared statement")
	}

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}

	// expect normal result
	mock.ExpectPrepare()
	stmt, err = db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}
	if stmt == nil {
		t.Errorf("stmt was expected while creating a prepared statement")
	}

	// expect error result
	mock.ExpectPrepare().WillReturnError(fmt.Errorf("Some DB error occurred"))
	stmt, err = db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err == nil {
		t.Error("error was expected while creating a prepared statement")
	}
	if stmt != nil {
		t.Errorf("stmt was not expected while creating a prepared statement returning error")
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestPreparedQueryExecutions(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs1 := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	rs2 := NewRows([]string{"id", "title"}).FromCSVString("2,whoop")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(2).
		WillReturnRows(rs2)

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected querying row from statement and scanning", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if title != "hello world" {
		t.Errorf("expected mocked title to be 'hello world', but got '%s' instead", title)
	}

	err = stmt.QueryRow(2).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected querying row from statement and scanning", err)
	}

	if id != 2 {
		t.Errorf("expected mocked id to be 2, but got %d instead", id)
	}

	if title != "whoop" {
		t.Errorf("expected mocked title to be 'whoop', but got '%s' instead", title)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestUnexpectedOperations(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected querying row, since there was no such expectation")
	}

	mock.ExpectRollback()

	err = db.Close()
	if err == nil {
		t.Error("error was expected while closing the database, expectation was not fulfilled", err)
	}
}

func TestWrongExpectations(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectBegin()

	rs1 := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	mock.ExpectCommit().WillReturnError(fmt.Errorf("deadlock occured"))
	mock.ExpectRollback() // won't be triggered

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ? FOR UPDATE")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected while querying row, since there begin transaction expectation is not fulfilled")
	}

	// lets go around and start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while querying row, since transaction was started", err)
	}

	err = tx.Commit()
	if err == nil {
		t.Error("a deadlock error was expected when commiting a transaction", err)
	}

	err = db.Close()
	if err == nil {
		t.Error("error was expected while closing the database, expectation was not fulfilled", err)
	}
}

func TestExecExpectations(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	result := NewResult(1, 1)
	mock.ExpectExec("^INSERT INTO articles").
		WithArgs("hello").
		WillReturnResult(result)

	res, err := db.Exec("INSERT INTO articles (title) VALUES (?)", "hello")
	if err != nil {
		t.Errorf("error '%s' was not expected, while inserting a row", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("error '%s' was not expected, while getting a last insert id", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		t.Errorf("error '%s' was not expected, while getting affected rows", err)
	}

	if id != 1 {
		t.Errorf("expected last insert id to be 1, but got %d instead", id)
	}

	if affected != 1 {
		t.Errorf("expected affected rows to be 1, but got %d instead", affected)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestRowBuilderAndNilTypes(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs := NewRows([]string{"id", "active", "created", "status"}).
		AddRow(1, true, time.Now(), 5).
		AddRow(2, false, nil, nil)

	mock.ExpectQuery("SELECT (.+) FROM sales").WillReturnRows(rs)

	rows, err := db.Query("SELECT * FROM sales")
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer rows.Close()

	// NullTime and NullInt are used from stubs_test.go
	var (
		id      int
		active  bool
		created NullTime
		status  NullInt
	)

	if !rows.Next() {
		t.Error("it must have had row in rows, but got empty result set instead")
	}

	err = rows.Scan(&id, &active, &created, &status)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 1 {
		t.Errorf("expected mocked id to be 1, but got %d instead", id)
	}

	if !active {
		t.Errorf("expected 'active' to be 'true', but got '%v' instead", active)
	}

	if !created.Valid {
		t.Errorf("expected 'created' to be valid, but it %+v is not", created)
	}

	if !status.Valid {
		t.Errorf("expected 'status' to be valid, but it %+v is not", status)
	}

	if status.Integer != 5 {
		t.Errorf("expected 'status' to be '5', but got '%d'", status.Integer)
	}

	// test second row
	if !rows.Next() {
		t.Error("it must have had row in rows, but got empty result set instead")
	}

	err = rows.Scan(&id, &active, &created, &status)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 2 {
		t.Errorf("expected mocked id to be 2, but got %d instead", id)
	}

	if active {
		t.Errorf("expected 'active' to be 'false', but got '%v' instead", active)
	}

	if created.Valid {
		t.Errorf("expected 'created' to be invalid, but it %+v is not", created)
	}

	if status.Valid {
		t.Errorf("expected 'status' to be invalid, but it %+v is not", status)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestArgumentReflectValueTypeError(t *testing.T) {
	mock, db, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs := NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery("SELECT (.+) FROM sales").WithArgs(5.5).WillReturnRows(rs)

	_, err = db.Query("SELECT * FROM sales WHERE x = ?", 5)
	if err == nil {
		t.Error("Expected error, but got none")
	}
}
