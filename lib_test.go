package scanstruct

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ory/dockertest"
	"log"
	"os"
	"testing"
)

var mySql *sql.DB

const doDockerStuff = true

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if doDockerStuff {
		if err := pool.Retry(func() error {
			var err error
			mySql, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
			if err != nil {
				return err
			}
			return mySql.Ping()
		}); err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}
	}

	code := m.Run()

	defer func() {
		if doDockerStuff {
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("Could not purge resource: %s", err)
			}
		}
		os.Exit(code)
	}()
}

func TestStupidSqlite(t *testing.T) {
	type Person struct {
		Age  int
		Name string
	}
	m := initTable(t, sqlite(t))
	person := Person{}
	err := ScanStruct(&person, m)
	if err != nil {
		t.Error(err)
	}
	if person.Age != 13 || person.Name != "Arne" {
		t.Error("Parsing did not compute.")
	}
}

func sqlite(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPointersSqlite(t *testing.T) {
	type Person struct {
		Age  *int
		Name *string
	}
	m := initTable(t, sqlite(t))
	var p Person
	err := ScanStruct(&p, m)
	if err != nil {
		t.Error(err)
	}
	if *p.Age != 13 || *p.Name != "Arne" {
		t.Error("parsed values did not work")
	}
}

func TestMySql(t *testing.T) {
	person := struct {
		Name string
		Age  int
	}{}
	m := initTable(t, mySql)
	err := ScanStruct(&person, m)
	if err != nil {
		t.Error(err)
	}
	if person.Age != 13 || person.Name != "Arne" {
		t.Error("Parsing did not compute.")
	}
}

func TestMysqlPointer(t *testing.T) {
	person := struct {
		Name *string
		Age  *int
	}{}
	m := initTable(t, mySql)
	err := ScanStruct(&person, m)
	if err != nil {
		t.Error(err)
	}
	if *person.Age != 13 || *person.Name != "Arne" {
		t.Error("Parsing did not compute.")
	}
}

func TestExtraColumns(t *testing.T) {
	person := struct {
		Age   int
		Name  string
		Extra int
	}{}
	m := initTable(t, sqlite(t))
	err := ScanStruct(&person, m)
	if err != nil {
		t.Error(err)
	}
}

func TestWrappedType(t *testing.T) {
	type Name string
	person := struct {
		Age int
		Name Name
	} {}
	m := initTable(t, sqlite(t))
	err := ScanStruct(&person, m)
	if err != nil {
		t.Error(err)
	}
	if person.Age != 13 || person.Name != "Arne" {
		t.Error("Parsing did not compute.")
	}
}

func initTable(t *testing.T, db *sql.DB) *sql.Rows {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS People(
		age int,
		name varchar(200)
	);`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
	INSERT INTO People(age, name) VALUES (13, 'Arne');
	`)
	if err != nil {
		t.Fatal(err)
	}

	m, err := db.Query("SELECT * FROM People")
	if err != nil {
		t.Fatal(err)
	}
	m.Next()
	return m
}
