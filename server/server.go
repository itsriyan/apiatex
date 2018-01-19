package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/riyan/apiatex/controllers"
	"github.com/riyan/apiatex/controllers/webserver/webhandler"
	"github.com/riyan/apiatex/models"
)

var (
	db        *sql.DB
	dbatex    = "atex"
	defaultdb = "mysql"
	dbuser    = "root"
	dbpass    = ""
	host      = ""
)

func main() {
	flag.Parse()
	var err error
	db, err := connectDB(dbatex, dbuser, dbpass, host)
	if err != nil {
		// if !isErrDBNotExist(err) {
		// 	log.Fatalf("Gagal Konek database %s", err)
		// }
		db, err = prepareDB()
		if err != nil {
			log.Fatal(err)
		}
	}
	webhandler.RegisterDB(db)
	webhandler.DebugOn()
	eurl := "/api/v1/e/"
	http.Handle(eurl, controllers.WebHandler(eurl))
	addr := ":8082"
	fmt.Printf("server start on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func connectDB(name, user, password, host string) (*sql.DB, error) {
	if password == "" {
		password = ":" + password
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s%s@%s/%s?parseTime=true", user, password, host, name))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	return db, err
}

// func isErrDBNotExist(err error) bool {
// 	if et, ok := err.(*.Error); ok {
// 		return et.Code == pq.ErrorCode("3D000")
// 	}
// 	return false
// }
func prepareDB() (*sql.DB, error) {
	// deleteDB()

	db, err := connectDB(defaultdb, dbuser, dbpass, host)
	if err != nil {
		return nil, err
	}

	if err = createDB(db, dbatex); err != nil {
		return nil, err
	}
	db.Close()
	if db, err = connectDB(dbatex, dbuser, dbpass, host); err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(models.EmployedTable); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return db, nil
}

func createDB(db *sql.DB, name string) error {
	query := "CREATE DATABASE " + name
	_, err := db.Exec(query)
	return err
}
