package models

import (
	"database/sql"
	"testing"
)

func TestEmployedCRUD(t *testing.T) {
	db := PrepareTest()
	defer db.Close()

	data := []*Employed{
		{
			Id:           "123456789",
			NameEmployed: "Tony Agus",
			Email:        "tonysetwan@atex.co.id",
			Phone:        "08521021312",
			Address:      "Cikupa aja sayyyyyyyyy",
		},
	}
	fields, _ := data[0].Fields()
	t.Run("InsertGet", func(t *testing.T) {
		//Insert
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range data {
			if err := v.Insert(tx); err != nil {
				t.Fatalf("insert  data:%v err:%v", v, err)
			}
			data[i] = v
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		//Get
		for _, v := range data {
			got := &Employed{Id: v.Id}
			err := got.Get(db)
			if err != nil {
				t.Fatalf("get data:%v err:%v", v, err)
			}
			compareDataEmployed(t, got, v, fields)
		}

		for _, v := range data {
			err := v.Insert(db)
			if err == nil {
				t.Error("insert duplicate data want error got nil")
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		// t.Parallel()
		td := data[0]
		changes := []map[string]interface{}{
			{"name_employed": "Updated description"},
			{"email": "K@kkk.com"},
			{"phone": "539583"},
		}
		for _, change := range changes {
			_, err := td.Update(db, change)
			if err != nil {
				t.Fatalf("update err:%v", err)
			}

			got := &Employed{Id: td.Id}
			err = got.Get(db)
			if err != nil {
				t.Fatal(err)
			}
			if v, ok := change["name_employed"]; ok {
				if got.NameEmployed != v {
					t.Errorf("name got:%s want:%s", got.NameEmployed, v)
				}
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// t.Parallel()
		for _, v := range data {
			if err := v.Delete(db); err != nil {
				t.Fatalf("delete td id:%s err:%v", v.Id, err)
			}
		}
		for _, v := range data {
			err := v.Get(db)
			if err == nil {
				t.Errorf("go err: nil want err:%v", sql.ErrNoRows)
			}
			if err != sql.ErrNoRows {
				t.Errorf("got err:%v want:%v", err, sql.ErrNoRows)
			}
		}
	})
}

func compareDataEmployed(t *testing.T, got, want *Employed, fields []string) {
	if len(fields) == 0 {
		fields, _ = got.Fields()
	}
	for _, field := range fields {
		if field == "id" && got.Id != want.Id {
			t.Errorf("got id:%s want:%s", got.Id, want.Id)
		}
		if field == "name_employed" && got.NameEmployed != want.NameEmployed {
			t.Errorf("got Name Employed:%s want:%s", got.NameEmployed, want.NameEmployed)
		}
		if field == "email" && got.Email != want.Email {
			t.Errorf("got email:%s want:%s", got.Email, want.Email)
		}
		if field == "phone" && got.Phone != want.Phone {
			t.Errorf("got phone:%s want:%s", got.Phone, want.Phone)
		}
		if field == "address" && got.Address != want.Address {
			t.Errorf("got address:%s want:%s", got.Address, want.Address)
		}
	}
}
