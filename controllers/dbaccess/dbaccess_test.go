package dbaccess

import (
	"database/sql"
	"log"
	"sort"
	"testing"
	"time"
)

var defaultdb, user, password, dbtest string

func init() {
	dbtest = "test_dbaccess"
	defaultdb = "mysql"
	user = "root"
	password = ""
	deleteDB()
}

var testTableSql = table{
	Name: "test_table",
	Fields: []string{
		"code VARCHAR(30) PRIMARY KEY",
		"description VARCHAR(100)",
		"transaction_date DATE",
		"amount NUMERIC",
		"count INTEGER",
	},
}

var testTableAutoSql = table{
	Name: "test_table_auto",
	Fields: []string{
		"code int AUTO_INCREMENT PRIMARY KEY",
		"description VARCHAR(100)",
		"transaction_date DATE",
		"amount NUMERIC",
		"count INTEGER",
	},
}

type testTable struct {
	Code            string    `json:"code"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	Amount          float64   `json:"amount"`
	Count           int       `json:"count"`
}

func (t *testTable) Name() string {
	return "test_table"
}

func (t *testTable) PrimaryKey() (fields []string, dst []interface{}) {
	fields = []string{"code"}
	dst = []interface{}{t.Code}
	return
}

func (t *testTable) HasAutoIncrementField() bool {
	return false
}

func (t *testTable) New() Table {
	return &testTable{}
}

func (t *testTable) Fields() (fields []string, dst []interface{}) {
	fields = []string{"code", "description", "transaction_date", "amount", "count"}
	dst = []interface{}{&t.Code, &t.Description, &t.TransactionDate, &t.Amount, &t.Count}
	return
}

type testTableAuto struct {
	Code            int       `json:"code"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	Amount          float64   `json:"amount"`
	Count           int       `json:"count"`
}

func (t *testTableAuto) Name() string {
	return "test_table_auto"
}

func (t *testTableAuto) PrimaryKey() (fields []string, dst []interface{}) {
	fields = []string{"code"}
	dst = []interface{}{t.Code}
	return
}

func (t *testTableAuto) HasAutoIncrementField() bool {
	return true
}

func (t *testTableAuto) New() Table {
	return &testTableAuto{}
}

func (t *testTableAuto) Fields() (fields []string, dst []interface{}) {
	fields = []string{"code", "description", "transaction_date", "amount", "count"}
	dst = []interface{}{&t.Code, &t.Description, &t.TransactionDate, &t.Amount, &t.Count}
	return
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

func prepareTest(t *testing.T) *sql.DB {
	deleteDB()
	db, err := connectDB(defaultdb, user, password, "")
	if err != nil {
		t.Fatal(err)
	}
	if err = createDB(db, dbtest); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if db, err = connectDB(dbtest, user, password, ""); err != nil {
		t.Fatal(err)
	}
	if err := createTable(db, testTableSql); err != nil {
		t.Fatal(err)
	}
	if err := createTable(db, testTableAutoSql); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCRUD(t *testing.T) {
	db := prepareTest(t)
	defer db.Close()
	nd := time.Now()
	now := time.Date(nd.Year(), nd.Month(), nd.Day(), 0, 0, 0, 0, time.UTC)
	data := []*testTable{
		{
			Code:            "KW",
			Description:     "Karawaci",
			TransactionDate: now,
			Amount:          100.00,
			Count:           1,
		},
	}
	fields, _ := data[0].Fields()
	t.Run("InsertGet", func(t *testing.T) {
		//Insert
		tx := newTx(t, db)
		for i, v := range data {
			if err := Insert(tx, v); err != nil {
				t.Fatalf("insert  data:%v err:%v", v, err)
			}
			data[i] = v
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		//Get
		for _, v := range data {
			got := &testTable{Code: v.Code}
			err := Get(db, got)
			if err != nil {
				t.Fatalf("get data:%v err:%v", v, err)
			}
			compareData(t, got, v, fields)
		}

		for _, v := range data {
			err := Insert(db, v)
			if err == nil {
				t.Error("insert duplicate data want error got nil")
			}
		}
	})

	t.Run("Exists", func(t *testing.T) {
		for _, v := range data {
			f := Filter{
				Field: "description",
				Op:    "=",
				Value: v.Description,
			}
			got, err := Exists(db, v, f)
			if err != nil {
				t.Fatal(err)
			}
			if !got {
				t.Errorf("got exist:%t want:true", got)
			}
		}
		f := Filter{
			Field: "description",
			Op:    "=",
			Value: "XXX should Not Exist",
		}
		got, err := Exists(db, data[0], f)
		if err != nil {
			t.Fatal(err)
		}
		if got {
			t.Errorf("got exist:%t want:false", got)
		}
	})

	t.Run("Update", func(t *testing.T) {
		// t.Parallel()
		td := data[0]
		changes := []map[string]interface{}{
			{"description": "Updated description"},
		}
		for _, change := range changes {
			if err := Update(db, td, change); err != nil {
				t.Fatalf("update err:%v", err)
			}
			got := &testTable{Code: td.Code}
			err := Get(db, got)
			if err != nil {
				t.Fatal(err)
			}
			if v, ok := change["description"]; ok {
				if got.Description != v {
					t.Errorf("name got:%s want:%s", got.Description, v)
				}
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// t.Parallel()
		for _, v := range data {
			if err := Delete(db, v); err != nil {
				t.Fatalf("delete td code:%s err:%v", v.Code, err)
			}
		}
		for _, v := range data {
			err := Get(db, v)
			if err == nil {
				t.Errorf("go err: nil want err:%v", sql.ErrNoRows)
			}
			if err != sql.ErrNoRows {
				t.Errorf("got err:%v want:%v", err, sql.ErrNoRows)
			}
		}
	})
}

func TestUpdateDeleteAll(t *testing.T) {
	db := prepareTest(t)
	defer db.Close()
	nd := time.Now()
	now := time.Date(nd.Year(), nd.Month(), nd.Day(), 0, 0, 0, 0, time.UTC)
	data := []*testTable{
		{
			Code:            "AA",
			Description:     "Karawaci",
			TransactionDate: now,
			Amount:          100.00,
			Count:           1,
		},
		{
			Code:            "AB",
			Description:     "Karawaci",
			TransactionDate: now,
			Amount:          100.00,
			Count:           2,
		},
		{
			Code:            "AC",
			Description:     "Karawaci",
			TransactionDate: now,
			Amount:          100.00,
			Count:           1,
		},
		{
			Code:            "AD",
			Description:     "Karawaci",
			TransactionDate: now,
			Amount:          100.00,
			Count:           2,
		},
	}

	//Insert
	tx := newTx(t, db)
	for i, v := range data {
		if err := Insert(tx, v); err != nil {
			t.Fatalf("insert  data:%v err:%v", v, err)
		}
		data[i] = v
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	t.Run("UpdateAll", func(t *testing.T) {
		// t.Parallel()
		testCase := []struct {
			fs             []Filter
			want           []*testTable
			newDescription string
		}{
			{
				newDescription: "UpdateAll no filter",
				want:           data[0:],
			},
			{
				fs: []Filter{
					{Field: "count", Op: "=", Value: 1},
				},
				newDescription: "UpdateAll no filter",
				want:           []*testTable{data[0], data[2]},
			},
		}
		td := data[0]
		for i, tc := range testCase {
			change := map[string]interface{}{
				"description": tc.newDescription,
			}
			if err := UpdateAll(db, td, change, tc.fs...); err != nil {
				t.Fatalf("update err:%v", err)
			}
			c := Cursor{
				Filters: tc.fs,
			}
			got, c, err := Fetch(db, td, c)
			if err != nil {
				t.Fatal(err)
			}
			fields := []string{"description"}
			for _, w := range tc.want {
				w.Description = tc.newDescription
			}
			compareResults(t, i, 0, got, tc.want, fields)
		}
	})

	t.Run("DeleteAll", func(t *testing.T) {
		// t.Parallel()
		testCase := []struct {
			fs   []Filter
			want []*testTable
		}{
			{
				fs: []Filter{
					{Field: "count", Op: "=", Value: 1},
				},
				want: []*testTable{data[1], data[3]},
			},
			{
				want: nil,
			},
		}
		td := data[0]
		fields, _ := td.Fields()
		for i, tc := range testCase {
			if err := DeleteAll(db, td, tc.fs...); err != nil {
				t.Fatalf("delete err:%v", err)
			}
			c := Cursor{}
			got, c, err := Fetch(db, td, c)
			if err != nil {
				t.Fatal(err)
			}
			compareResults(t, i, 0, got, tc.want, fields)
		}
	})

}

func TestFetchLimit(t *testing.T) {
	db := prepareTest(t)
	defer db.Close()
	nd := time.Now()
	now := time.Date(nd.Year(), nd.Month(), nd.Day(), 0, 0, 0, 0, time.UTC)
	data := []*testTable{
		{Code: "AA", Description: "AA Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "AB", Description: "AB Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "AC", Description: "AC Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BA", Description: "BA Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BB", Description: "BB Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BC", Description: "BC Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BD", Description: "BD Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
	}
	tx := newTx(t, db)
	for i, v := range data {
		if err := Insert(tx, v); err != nil {
			t.Fatalf("insert  data:%v err:%v", v, err)
		}
		data[i] = v
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	testCase := []struct {
		c     Cursor
		wants [][]*testTable
	}{
		{
			c: Cursor{
				Limit: 0,
			},
			wants: [][]*testTable{
				data[0:],
			},
		},
		{
			c: Cursor{
				Limit: 2,
			},
			wants: [][]*testTable{
				data[:2],
				data[2:4],
				data[4:6],
				data[6:],
			},
		},
		{
			c: Cursor{
				Limit:      2,
				Descending: true,
			},
			wants: [][]*testTable{
				[]*testTable{data[6], data[5]},
				[]*testTable{data[4], data[3]},
				[]*testTable{data[2], data[1]},
				[]*testTable{data[0]},
			},
		},
		{
			c: Cursor{
				Fields: []string{"code", "description"},
				Limit:  2,
			},
			wants: [][]*testTable{
				data[:2],
				data[2:4],
				data[4:6],
				data[6:],
			},
		},
	}
	var (
		got []Table
		err error
	)
	for i, tc := range testCase {
		c := tc.c
		for j, want := range tc.wants {
			got, c, err = Fetch(db, data[0], c)
			if err != nil {
				t.Fatal(err)
			}
			// for _, g := range got {
			// 	fmt.Printf("tc:%d td:%d got:%v\n", i, j, *g.(*testTable))
			// }
			compareResults(t, i, j, got, want, tc.c.Fields)
		}
	}
}

func TestFetchFilter(t *testing.T) {
	db := prepareTest(t)
	defer db.Close()
	nd := time.Now()
	now := time.Date(nd.Year(), nd.Month(), nd.Day(), 0, 0, 0, 0, time.UTC)
	data := []*testTable{
		{Code: "AA", Description: "AA Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "AB", Description: "AB Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "AC", Description: "AC Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BA", Description: "BA Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BB", Description: "BB Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BC", Description: "BC Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
		{Code: "BD", Description: "BD Karawaci", TransactionDate: now, Amount: 100.00, Count: 1},
	}
	tx := newTx(t, db)
	for i, v := range data {
		if err := Insert(tx, v); err != nil {
			t.Fatalf("insert  data:%v err:%v", v, err)
		}
		data[i] = v
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	testCase := []struct {
		c     Cursor
		wants [][]*testTable
	}{
		{
			c: Cursor{
				Filters: []Filter{
					{Field: "code", Op: "=", Value: "AC"},
				},
			},
			wants: [][]*testTable{
				data[2:3],
			},
		},
		{
			c: Cursor{
				Limit: 2,
			},
			wants: [][]*testTable{
				data[:2],
				data[2:4],
				data[4:6],
				data[6:],
			},
		},
		{
			c: Cursor{
				Limit:      2,
				Descending: true,
			},
			wants: [][]*testTable{
				[]*testTable{data[6], data[5]},
				[]*testTable{data[4], data[3]},
				[]*testTable{data[2], data[1]},
				[]*testTable{data[0]},
			},
		},
		{
			c: Cursor{
				Fields: []string{"code", "description"},
				Limit:  2,
			},
			wants: [][]*testTable{
				data[:2],
				data[2:4],
				data[4:6],
				data[6:],
			},
		},
	}
	var (
		got []Table
		err error
	)
	for i, tc := range testCase {
		c := tc.c
		for j, want := range tc.wants {
			got, c, err = Fetch(db, data[0], c)
			if err != nil {
				t.Fatal(err)
			}
			// for _, g := range got {
			// 	fmt.Printf("tc:%d td:%d got:%v\n", i, j, *g.(*testTable))
			// }
			compareResults(t, i, j, got, want, tc.c.Fields)
		}
	}
}

func compareResults(t *testing.T, tc, td int, got []Table, want []*testTable, fields []string) {
	if len(got) != len(want) {
		t.Fatalf("tc:%d td:%d data got:%d want:%d", tc, td, len(got), len(want))
	}
	for i, g := range got {
		compareData(t, g.(*testTable), want[i], fields)
	}
}

func compareData(t *testing.T, got, want *testTable, fields []string) {
	if len(fields) == 0 {
		fields, _ = got.Fields()
	}
	for _, field := range fields {
		if field == "code" && got.Code != want.Code {
			t.Errorf("got code:%s want:%s", got.Code, want.Code)
		}
		if field == "description" && got.Description != want.Description {
			t.Errorf("got description:%s want:%s", got.Description, want.Description)
		}
		if field == "transaction_date" && !got.TransactionDate.Equal(want.TransactionDate) {
			t.Errorf("got trancation date:%s want:%s", got.TransactionDate, want.TransactionDate)
		}
		if field == "amount" && got.Amount != want.Amount {
			t.Errorf("got amount:%.5f want:%.5f", got.Amount, want.Amount)
		}
		if field == "count" && got.Count != want.Count {
			t.Errorf("got count:%d want:%d", got.Count, want.Count)
		}
	}
}

func newTx(t *testing.T, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func sortByCode(data []*testTable) []*testTable {
	byCode := func(i, j int) bool {
		return data[i].Code < data[j].Code
	}
	sort.Slice(data, byCode)
	return data
}

func sortTableByCode(data []Table) []*testTable {
	byCode := func(i, j int) bool {
		return data[i].(*testTable).Code < data[j].(*testTable).Code
	}
	sort.Slice(data, byCode)
	result := make([]*testTable, len(data))
	for i, v := range data {
		result[i] = v.(*testTable)
	}
	return result
}

// func TestAutoIncrementCRUD(t *testing.T) {
// 	db := prepareTest(t)
// 	defer db.Close()
// 	nd := time.Now()
// 	now := time.Date(nd.Year(), nd.Month(), nd.Day(), 0, 0, 0, 0, time.UTC)
// 	data := []*testTableAuto{
// 		{
// 			Description:     "Karawaci",
// 			TransactionDate: now,
// 			Amount:          100.00,
// 			Count:           1,
// 		},
// 	}
// 	t.Run("InsertGet", func(t *testing.T) {
// 		//Insert
// 		tx := newTx(t, db)
// 		for i, v := range data {
// 			if err := Insert(tx, v); err != nil {
// 				t.Fatalf("insert  data:%v err:%v", v, err)
// 			}
// 			data[i] = v
// 		}
// 		if err := tx.Commit(); err != nil {
// 			t.Fatal(err)
// 		}
//
// 		//Get
// 		for _, v := range data {
// 			got := &testTableAuto{Code: v.Code}
// 			err := Get(db, got)
// 			if err != nil {
// 				t.Fatalf("get data:%v err:%v", v, err)
// 			}
// 			compareAutoData(t, got, v)
// 		}
// 	})
//
// 	t.Run("Update", func(t *testing.T) {
// 		// t.Parallel()
// 		td := data[0]
// 		changes := []map[string]interface{}{
// 			{"description": "Updated description"},
// 		}
// 		for _, change := range changes {
// 			if err := Update(db, td, change); err != nil {
// 				t.Fatalf("update err:%v", err)
// 			}
// 			got := &testTableAuto{Code: td.Code}
// 			err := Get(db, got)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if v, ok := change["description"]; ok {
// 				if got.Description != v {
// 					t.Errorf("name got:%s want:%s", got.Description, v)
// 				}
// 			}
// 		}
// 	})
//
// 	t.Run("Delete", func(t *testing.T) {
// 		// t.Parallel()
// 		for _, v := range data {
// 			if err := Delete(db, v); err != nil {
// 				t.Fatalf("delete td code:%d err:%v", v.Code, err)
// 			}
// 		}
// 		for _, v := range data {
// 			err := Get(db, v)
// 			if err == nil {
// 				t.Errorf("go err: nil want err:%v", sql.ErrNoRows)
// 			}
// 			if err != sql.ErrNoRows {
// 				t.Errorf("got err:%v want:%v", err, sql.ErrNoRows)
// 			}
// 		}
// 	})
// }

func compareAutoData(t *testing.T, got, want *testTableAuto) {
	if got.Code != want.Code {
		t.Errorf("got code:%d want:%d", got.Code, want.Code)
	}
	if got.Description != want.Description {
		t.Errorf("got description:%s want:%s", got.Description, want.Description)
	}
	if !got.TransactionDate.Equal(want.TransactionDate) {
		t.Errorf("got trancation date:%s want:%s", got.TransactionDate, want.TransactionDate)
	}
	if got.Amount != want.Amount {
		t.Errorf("got amount:%.5f want:%.5f", got.Amount, want.Amount)
	}
	if got.Count != want.Count {
		t.Errorf("got count:%d want:%d", got.Count, want.Count)
	}
}

func sortAutoByCode(data []*testTableAuto) []*testTableAuto {
	byCode := func(i, j int) bool {
		return data[i].Code < data[j].Code
	}
	sort.Slice(data, byCode)
	return data
}

func sortAutoTableByCode(data []Table) []*testTableAuto {
	byCode := func(i, j int) bool {
		return data[i].(*testTableAuto).Code < data[j].(*testTableAuto).Code
	}
	sort.Slice(data, byCode)
	result := make([]*testTableAuto, len(data))
	for i, v := range data {
		result[i] = v.(*testTableAuto)
	}
	return result
}
