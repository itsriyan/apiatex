package models

import (
	"github.com/riyan/apiatex/controllers/dbaccess"
)

var EmployedTable = `CREATE TABLE employed
(
		id varchar(10) PRIMARY KEY,
    name_employed varchar(100) not null,
    email varchar(100) not null,
    phone varchar(13) not null,
    address varchar(400)
	);`

type Employed struct {
	Id           string `json:"id"`
	NameEmployed string `json:"name_employed"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
}

//Name resturn nama table
func (e *Employed) Name() string {
	return "employed"
}

//PrimaryKey return PrimaryKey table
func (em *Employed) PrimaryKey() (fields []string, dst []interface{}) {
	fields = []string{"id"}
	dst = []interface{}{&em.Id}
	return fields, dst
}

//New Membuat baru
func (me *Employed) New() dbaccess.Table {
	return &Employed{}
}

// Fields Deklarei coloumn yang ada di eset
func (em *Employed) Fields() (fields []string, dst []interface{}) {
	fields = []string{"id", "name_employed", "email", "phone", "address"}
	dst = []interface{}{&em.Id, &em.NameEmployed, &em.Email, &em.Phone, &em.Address}
	return fields, dst
}

//HeAutoIncrementField false karna tidak ada AUTO NUMBER/SERIAL
func (em *Employed) HasAutoIncrementField() bool {
	return false
}

//Insert Untuk fungsi memeukan eset
func (em *Employed) Insert(db dbaccess.DBExecer) error {
	return dbaccess.Insert(db, em)
}

// Update fungsi mengedit data eset
func (em *Employed) Update(db dbaccess.DBExecer, change map[string]interface{}) (map[string]interface{}, error) {
	err := dbaccess.Update(db, em, change)
	return change, err
}

//Delete fungsi mengahapus data eset
func (em *Employed) Delete(db dbaccess.DBExecer) error {
	return dbaccess.Delete(db, em)
}

//Get fungsi untuk mengambil data eset
func (em *Employed) Get(db dbaccess.DBExecer) error {
	return dbaccess.Get(db, em)
}
