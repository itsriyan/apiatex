package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var defaultdb, user, password, dbtest string

func init() {
	dbtest = "test_dbaccess"
	defaultdb = "mysql"
	user = "root"
	password = ""
	deleteDB()
}

//connectDB connect to the database.
func connectDB(dbname, user, password, host string) (*sql.DB, error) {
	if password == "" {
		password = ":" + password
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s%s@%s/%s?parseTime=true", user, password, host, dbname))
	if err != nil {
		return nil, err
	}
	return db, nil
}

//createDB create new database.
func createDB(db *sql.DB, name string) error {
	query := "CREATE DATABASE " + name
	_, err := db.Exec(query)
	return err
}

//dropDB delete the database.
func dropDB(db *sql.DB, name string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	_, err := db.Exec(query)
	return err
}

//Table adalah
type table struct {
	Name       string
	Fields     []string
	PrimaryKey string
}

//createTable based
func createTable(db *sql.DB, table table) error {
	var query string
	if table.PrimaryKey != "" {
		query = fmt.Sprintf("CREATE TABLE %s (%s, PRIMARY KEY %s)", table.Name,
			strings.Join(table.Fields, ","), table.PrimaryKey)
	} else {
		query = fmt.Sprintf("CREATE TABLE %s (%s)", table.Name, strings.Join(table.Fields, ","))
	}
	_, err := db.Exec(query)
	return err
}
func deleteDB() {
	//DeleteDatabase if exist
	db, err := connectDB(defaultdb, user, password, "")
	if err != nil {
		log.Fatal(err)
	}
	if err := dropDB(db, dbtest); err != nil {
		log.Fatal(err)
	}
}
func PrepareTest() *sql.DB {
	deleteDB()
	db, err := connectDB(defaultdb, user, password, "")
	if err != nil {
		log.Fatal(err)
	}
	if err = createDB(db, dbtest); err != nil {
		log.Fatal(err)
	}
	db.Close()
	if db, err = connectDB(dbtest, user, password, ""); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(EmployedTable); err != nil {
		log.Fatal(err)
	}
	// for _, ne := range number {
	// 	if err := ne.Insert(db); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	return db
}
