// Package dbaccess provide helper for the CRUD operation on database.
package dbaccess

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

//DBExecer is an interface that can run query or execute tot the database.
//*sql.DB, *sql.Stmt and *sql.Tx implement DBExecer.
type DBExecer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//Table ...
type Table interface {
	///Name return the name of the table on the database.
	Name() string
	//PrimaryKey return the PrimaryKey field.
	PrimaryKey() (fields []string, dst []interface{})
	New() Table
	//Fields return the name of the field that match with the name of the field
	//on the database and dst is fields that will store the value when scan data
	//from the database, dst must be pointer.
	//Fields include all the fields on PrimaryKey.
	Fields() (fields []string, dst []interface{})
	//AutoIncrementFieldIdx return true if there is auto increment field,
	//and on the fields auto increment field should be on the first.
	HasAutoIncrementField() bool
}

type Filterer interface {
	//Filter return string to put as where.
	// n is the placeholder number.
	Filter(n int) (string, interface{})
}

//Insert data from table to the database.
func Insert(db DBExecer, t Table) error {
	if t.HasAutoIncrementField() {
		return insertAutoIncr(db, t)
	}
	fields, dst := t.Fields()
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.Name(), strings.Join(fields, ","), makePlaceHolder(1, len(dst)))
	_, err := db.Exec(query, dst...)
	return err
}

//Exists return true if there is a record match with the filters.
func Exists(db DBExecer, t Table, filters ...Filter) (bool, error) {
	where, args, err := whereClause(t, 1, filters...)
	if err != nil {
		return false, err
	}
	query := fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE %s)", t.Name(), where)
	exist := false
	err = db.QueryRow(query, args...).Scan(&exist)
	return exist, err
}

func insertAutoIncr(db DBExecer, t Table) error {
	fields, dst := t.Fields()
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		t.Name(), strings.Join(fields[1:], ","), makePlaceHolder(1, len(dst)-1))
	_, err := db.Exec(query, dst[1:]...)
	ssc := fmt.Sprintf("SELECT max(%s) from %s", fields[0], t.Name())
	err = db.QueryRow(ssc).Scan(&dst[0])
	return err
}

//Update update the data on the database with the change.
func Update(db DBExecer, t Table, change map[string]interface{}) error {
	set, args, err := setQuery(t, change)
	if err != nil {
		return err
	}
	fields, dst := t.PrimaryKey()
	w := where(fields)
	args = append(args, dst...)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", t.Name(), set, w)
	_, err = db.Exec(query, args...)
	return err
}

func setQuery(t Table, change map[string]interface{}) (set string, args []interface{}, err error) {
	if len(change) == 0 {
		err = errors.New("no changes to update")
		return set, args, err
	}
	var sets []string
	for k, v := range change {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		args = append(args, v)

	}
	set = strings.Join(sets, ",")
	return set, args, err
}

//DeleteAll delete all the records mathc with the filters,
//if there is no filters, this will delete all the data from the table.
func UpdateAll(db DBExecer, t Table, change map[string]interface{}, fs ...Filter) error {
	set, args, err := setQuery(t, change)
	if err != nil {
		return err
	}
	count := len(change) + 1
	c := Cursor{
		Filters: fs,
	}
	where, wargs, err := filters(t, c, count)
	if err != nil {
		return err
	}
	args = append(args, wargs...)
	var query string
	if where != "" {
		query = fmt.Sprintf("UPDATE %s SET %s WHERE %s", t.Name(), set, where)
	} else {
		query = fmt.Sprintf("UPDATE %s SET %s", t.Name(), set)
	}
	_, err = db.Exec(query, args...)
	return err
}

//Fetch get all the records that match with the cursor.
//Zero value cursor mean fetch will get all the records from the database.
//Fetch support paging using limint on Cursor, and return the Cursor to get the nex records set.
func Fetch(db DBExecer, t Table, option Cursor) (result []Table, c Cursor, err error) {
	if len(option.Fields) == 0 {
		option.Fields, _ = t.Fields()
	}
	if err = isFieldsOK(t, option.Fields); err != nil {
		return result, option, err
	}
	if err = isFieldsOK(t, option.OrderBy); err != nil {
		return result, option, err
	}
	option.OrderBy = addUniqueField(t, option.OrderBy)
	query, whereArgs, err := queryFromCursor(t, option)
	if err != nil {
		return nil, option, err
	}
	// fmt.Printf("query:%s\n", query)
	// for _, v := range option.LastArgs {
	// 	fmt.Printf("%s\n", *v.(*string))
	// }
	// fmt.Println(option.LastArgs...)
	rows, err := db.Query(query, append(whereArgs, option.LastArgs...)...)
	if err != nil {
		return nil, option, err
	}
	result, c, err = scanRows(t, rows, option)
	return result, c, err
}

//scanRows scan the rows and close it.
func scanRows(t Table, rows *sql.Rows, c Cursor) (result []Table, cursor Cursor, err error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return result, c, err
	}
	cursor = c
	var last Table
	for rows.Next() {
		tbl := t.New()
		fields, dst := tbl.Fields()
		if len(fields) != len(cols) {
			dst, err = scanArgs(tbl, cols)
			if err != nil {
				return result, cursor, err
			}
		}
		if err = rows.Scan(dst...); err != nil {
			return nil, cursor, err
		}
		result = append(result, tbl)
		last = tbl
	}
	if err = rows.Err(); err != nil {
		return result, cursor, err
	}
	if last != nil {
		cursor.LastArgs, err = scanArgs(last, cursor.OrderBy)
	}
	return result, cursor, err
}

func queryFromCursor(t Table, c Cursor) (string, []interface{}, error) {
	where, queryArgs, err := filters(t, c, 1)
	if err != nil {
		return "", nil, err
	}
	var limit string
	if c.Limit != 0 {
		limit = fmt.Sprintf("LIMIT %d ", c.Limit)
	}
	var query string
	if where == "" {
		if c.Descending {
			query = fmt.Sprintf("SELECT %s FROM %s ORDER BY (%s) DESC %s",
				strings.Join(c.Fields, ","), t.Name(), strings.Join(c.OrderBy, ","), limit)
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s ORDER BY (%s) %s",
				strings.Join(c.Fields, ","), t.Name(), strings.Join(c.OrderBy, ","), limit)
		}
	} else {
		if c.Descending {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY (%s) DESC %s",
				strings.Join(c.Fields, ","), t.Name(), where, strings.Join(c.OrderBy, ","), limit)
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY (%s) %s",
				strings.Join(c.Fields, ","), t.Name(), where, strings.Join(c.OrderBy, ","), limit)
		}
	}
	// fmt.Printf("query:%s\n", query)
	return query, queryArgs, nil
}

func filters(t Table, c Cursor, start int) (where string, queryArgs []interface{}, err error) {
	where, queryArgs, err = whereClause(t, start, c.Filters...)
	if err != nil {
		return where, queryArgs, err
	}
	start += len(c.Filters)
	if len(c.LastArgs) != 0 {
		desc := ">"
		if c.Descending {
			desc = "<"
		}
		if where == "" {
			where = fmt.Sprintf("(%s) %s (%s)", strings.Join(c.OrderBy, ","), desc,
				makePlaceHolder(start, len(c.OrderBy)))
		} else {
			where = fmt.Sprintf("%s AND (%s) %s (%s)", where, strings.Join(c.OrderBy, ","), desc,
				makePlaceHolder(start, len(c.OrderBy)))
		}
	}
	return where, queryArgs, err
}

func whereClause(t Table, start int, filters ...Filter) (where string, queryArgs []interface{}, err error) {
	tblFields, _ := t.Fields()
	for _, filter := range filters {
		if err = filter.IsValidOp(); err != nil {
			return where, queryArgs, err
		}
		if _, exist := fieldExist(tblFields, filter.Field); !exist {
			err = fmt.Errorf("table:%s does not have field:%s", t.Name(), filter.Field)
			return where, queryArgs, err
		}
		if where != "" {
			where = fmt.Sprintf("%s AND %s %s ?", where, filter.Field, filter.Op)
		} else {
			where = fmt.Sprintf("%s %s ?", filter.Field, filter.Op)
		}
		queryArgs = append(queryArgs, filter.Value)
		start++
	}
	return where, queryArgs, err
}

//addUniqueField add the PrimaryKey to the OrderBy if there is no uniqure field
//on OrderBy.
func addUniqueField(t Table, fields []string) []string {
	pkFields, _ := t.PrimaryKey()
	if len(fields) == 0 {
		return pkFields
	}
	for _, pkf := range pkFields {
		if _, exist := fieldExist(fields, pkf); exist {
			return fields
		}
	}
	fields = append(fields, pkFields...)
	return fields
}

func isFieldsOK(t Table, fields []string) error {
	tblFields, _ := t.Fields()
	for _, field := range fields {
		if _, exist := fieldExist(tblFields, field); !exist {
			return fmt.Errorf("table:%s does not have field:%s", t.Name(), field)
		}
	}
	return nil
}

func fieldExist(fields []string, name string) (index int, exist bool) {
	for i, field := range fields {
		if field == name {
			index = i
			exist = true
			return index, exist
		}
	}
	return index, exist
}

//ScanArgs return the args for the fields, args is pointer to the instance that hold the
//Table.
func scanArgs(t Table, fields []string) (args []interface{}, err error) {
	tblFields, tblArgs := t.Fields()
	for _, field := range fields {
		index, exist := fieldExist(tblFields, field)
		if !exist {
			err = fmt.Errorf("table:%s does not have field:%s", t.Name(), field)
			return args, err
		}
		args = append(args, tblArgs[index])
	}
	return args, err
}

//Delete delete one record from the database that match with the PrimaryKey value.
func Delete(db DBExecer, t Table) error {
	fields, dst := t.PrimaryKey()
	w := where(fields)
	query := fmt.Sprintf("DELETE FROM %s where %s", t.Name(), w)
	_, err := db.Exec(query, dst...)
	return err
}

//DeleteAll delete all the records mathc with the filters,
//if there is no filters, this will delete all the data from the table.
func DeleteAll(db DBExecer, t Table, fs ...Filter) error {
	if len(fs) == 0 {
		_, err := db.Exec("DELETE FROM " + t.Name())
		return err
	}
	c := Cursor{
		Filters: fs,
	}

	where, args, err := filters(t, c, 1)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", t.Name(), where)
	// fmt.Println(query)
	_, err = db.Exec(query, args...)
	return err
}

//Get get one record from the database that match with table PrimaryKey value.
func Get(db DBExecer, t Table) error {
	fields, dst := t.Fields()
	pkf, pkd := t.PrimaryKey()
	w := where(pkf)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(fields, ","), t.Name(), w)
	err := db.QueryRow(query, pkd...).Scan(dst...)
	return err
}

func where(fields []string) string {
	var w []string
	for _, v := range fields {
		w = append(w, fmt.Sprintf("%s = ?", v))

	}
	return fmt.Sprintf(strings.Join(w, " AND "))
}

func makePlaceHolder(start, n int) string {
	r := make([]string, n)
	for i := range r {
		r[i] = fmt.Sprintf("?")
		start++
	}
	return strings.Join(r, ",")
}
