package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"

	"github.com/riyan/apiatex/controllers/dbaccess"
	"github.com/riyan/apiatex/controllers/webserver/webhandler"
	"github.com/riyan/apiatex/models"
)

// dbh provide db to handler.
type dbh struct {
	db *sql.DB
}

func (d dbh) Handle(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	res.Ctx = webhandler.NewContextWithDB(r.Context(), d.db)
	return res
}

var em = []*models.Employed{
	&models.Employed{
		Id: "11111111", NameEmployed: "Tony Agus", Email: "Tony@gmail.com",
		Phone: "0812312312", Address: "Kp. Cilengsiii",
	},
	&models.Employed{
		Id: "22222222", NameEmployed: "Agus Tony", Email: "Agus@gmail.com",
		Phone: "0812312312", Address: "Kp. Cocoloking",
	},
	&models.Employed{
		Id: "33333333", NameEmployed: "hawking", Email: "hawking@gmail.com",
		Phone: "0812312312", Address: "Kp. cicalangkang",
	},
	&models.Employed{
		Id: "44444444", NameEmployed: "John Doe", Email: "johndoe@gmail.com",
		Phone: "0812312312", Address: "Kp. America Latin",
	},
	&models.Employed{
		Id: "55555555", NameEmployed: "Doe John", Email: "Doe John@gmail.com",
		Phone: "0812312312", Address: "Kp. Lorem Ipsum Dolor sit amet",
	},
}

func TestEmployedHandler(t *testing.T) {
	dbt := PrepareTest()
	defer dbt.Close()
	handle := webhandler.New(dbh{db: dbt}, eHandler{pattern: "/api/v1/e/"})
	ts := httptest.NewServer(handle)
	defer ts.Close()

	t.Run("Get Request ALL Employed", func(t *testing.T) {
		urls := ts.URL + "/api/v1/e/employed/"
		res, err := http.Get(urls)
		if err != nil {
			t.Fatalf("get err:%v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("read response body err :%v", err)
		}

		type trd struct {
			Err    string            `json:"err"`
			Data   []models.Employed `json:"data"`
			Cursor dbaccess.Cursor   `json:"cursor"`
		}
		var rd trd

		if err := json.Unmarshal(data, &rd); err != nil {
			t.Fatalf("err : %v data: %s", err, data)
		}
		if rd.Err != "" {
			t.Fatalf("Read Response Error : %s", rd.Err)
		}
		employed := rd.Data
		if len(employed) != len(em) {
			t.Errorf("got data: %d want :%d\n", len(employed), len(em))
		}
		//fmt.Printf("result :%v\n", asset)
	})
	t.Run("GET 1 Employed", func(t *testing.T) {
		getone := em[1]
		urls := ts.URL + "/api/v1/e/employed/" + getone.Id
		res, err := http.Get(urls)
		if err != nil {

			t.Fatalf("get err:%v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("read response body err :%v", err)
		}

		type trd struct {
			Err    string          `json:"err"`
			Data   models.Employed `json:"data"`
			Cursor dbaccess.Cursor `json:"cursor"`
		}

		var rd trd

		if err = json.Unmarshal(data, &rd); err != nil {
			t.Fatalf("err : %v data: %s", err, data)
		}
		if rd.Err != "" {
			t.Fatalf("Read Response Error : %s", rd.Err)
		}
		employed := rd.Data
		if employed.NameEmployed != getone.NameEmployed {
			t.Errorf("got data: %v want :%v\n", employed.NameEmployed, getone.NameEmployed)
		}

	})
	t.Run("Filter Employed", func(t *testing.T) {
		urls := ts.URL + "/api/v1/e/employed/"
		var testemplyed = []struct {
			fil  string
			want []*models.Employed
		}{
			{
				fil:  "id,=,11111111",
				want: em[0:1],
			},
			{
				fil:  fmt.Sprintf("name_employed,like,%s", url.PathEscape("%%hawk%%")),
				want: em[2:3],
			},
		}
		for c, tc := range testemplyed {
			req, err := http.NewRequest("GET", urls+fmt.Sprintf("?filters=%s", url.QueryEscape(tc.fil)), nil)
			if err != nil {
				t.Fatalf("error get : %v", err)
			}
			clint := http.Client{}
			res, err := clint.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Printf("ini datanya : %s\n", body)
			type trd struct {
				Err    string             `json:"err"`
				Data   []*models.Employed `json:"data"`
				cursor dbaccess.Cursor    `json:"cursor"`
			}
			var rd trd
			if err := json.Unmarshal(body, &rd); err != nil {
				t.Fatalf("err : %v data: %s", err, body)
			}
			if rd.Err != "" {
				t.Fatal(rd.Err)
			}
			got := rd.Data
			if len(got) != len(tc.want) {
				t.Fatalf("data :%d  Len got : %d, len want : %d", c+1, len(got), len(tc.want))
			}
			g := UrutEmployed(got)
			c := UrutEmployed(tc.want)
			for i, v := range c {
				compareEmployed(t, g[i], v)
			}
		}
	})
	t.Run("Insert Employed", func(t *testing.T) {
		dat := models.Employed{
			Id: "66666666", NameEmployed: "Lorem Ipsum", Email: "Sit Amet@gmail.com",
			Phone: "0812312312", Address: "Kp. Lorem Ipsum Dolor sit amet",
		}
		b, err := json.MarshalIndent(dat, "", " ")
		if err != nil {
		}
		urls := ts.URL + "/api/v1/e/employed"
		res, err := http.Post(urls, "application/x-www-form-urlencoded", bytes.NewBuffer(b))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		type trd struct {
			Err  string           `json:"err"`
			Data *models.Employed `json:"data"`
		}
		var rd trd
		if err := json.Unmarshal(body, &rd); err != nil {
			t.Fatalf("err : %v data: %s", err, body)
		}
		if rd.Err != "" {
			t.Fatal(rd.Err)
		}
		gotrd := rd.Data
		if gotrd.Id != dat.Id {
			t.Errorf("get data:%v want:%v", gotrd, dat)
		}
		got := &models.Employed{Id: dat.Id}
		err = dbaccess.Get(dbt, got)
		if err != nil {
			t.Fatalf("get data:%v err:%v", dat, err)
		}
		if got.NameEmployed != dat.NameEmployed {
			t.Errorf("got : %s want : %s", got.NameEmployed, dat.NameEmployed)
		}
	})
	t.Run("Update Employed", func(t *testing.T) {
		datper := em[0]
		dataa := map[string]interface{}{
			"id": datper.Id, "name_employed": "Superman",
		}
		b, err := json.MarshalIndent(dataa, "", " ")
		if err != nil {
		}
		urls := ts.URL + "/api/v1/e/employed/" + fmt.Sprint(dataa["id"])
		// res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(b))
		req, err := http.NewRequest("PUT", urls, bytes.NewBuffer(b))
		if err != nil {
			t.Fatal(err)
		}
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		type trd struct {
			Err  string           `json:"err"`
			Data *models.Employed `json:"data"`
		}
		var rd trd
		if err := json.Unmarshal(body, &rd); err != nil {
			t.Fatalf("err : %v data: %s", err, body)
		}
		if rd.Err != "" {

		}
		gotrd := rd.Data
		if gotrd.Id != dataa["id"] {
			t.Errorf("get data:%v want:%v", gotrd, dataa)
		}
		got := &models.Employed{Id: datper.Id}
		err = dbaccess.Get(dbt, got)
		if err != nil {
			t.Fatalf("get data:%v err:%v", datper, err)
		}

	})
	t.Run("Delete Asset", func(t *testing.T) {

		employed := models.Employed{
			Id: "66666666",
		}
		urls := ts.URL + "/api/v1/e/employed/" + employed.Id

		req, err := http.NewRequest("DELETE", urls, nil)
		if err != nil {
			t.Fatal(err)
		}

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got status :%v Want:StatusOK", res.StatusCode)
		}
		defer res.Body.Close()

		got := &models.Employed{Id: employed.Id}
		err = dbaccess.Get(dbt, got)
		if err != nil {

		} else {
			t.Fatalf("get data:%v err: Tidak Terhapus", employed)
		}
	})
}

func UrutEmployed(data []*models.Employed) []*models.Employed {

	sort.Slice(data, func(i, j int) bool { return data[i].Id < data[j].Id })
	return data
}
func compareEmployed(t *testing.T, got, want *models.Employed) {

	if got.Id != want.Id {
		t.Errorf("got id:%s want:%s", got.Id, want.Id)
	}
	if got.NameEmployed != want.NameEmployed {
		t.Errorf("got Name Employe:%s want:%s", got.NameEmployed, want.NameEmployed)
	}
	if got.Email != want.Email {
		t.Errorf("got Email:%s want:%s", got.Email, want.Email)
	}
	if got.Phone != want.Phone {
		t.Errorf("got Phone:%s want:%s", got.Phone, want.Phone)
	}
	if got.Address != want.Address {
		t.Errorf("got Address:%s want:%s", got.Address, want.Address)
	}
}
