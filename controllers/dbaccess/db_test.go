package dbaccess

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

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
