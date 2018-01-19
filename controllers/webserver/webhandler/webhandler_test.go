package webhandler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockWH struct {
	msg            string
	handled        bool
	db             *sql.DB
	checkDbFromCtx bool
}

func (wh mockWH) Handle(w http.ResponseWriter, r *http.Request) Response {
	res := Response{
		Data: wh.msg,
	}
	if wh.db != nil {
		res.Ctx = NewContextWithDB(r.Context(), wh.db)
	}
	if wh.checkDbFromCtx {
		db, err := DBFromContext(r.Context())
		if err != nil {
			res.Error(err, http.StatusInternalServerError)
			return res
		}
		if db == nil {
			res.Error(errors.New("db is nil"), http.StatusInternalServerError)
			return res
		}
	}
	if wh.handled {
		res.Handled = true
		fmt.Fprintf(w, "%s", wh.msg)
	}
	return res
}

func TestHandler(t *testing.T) {
	setGlogFlag()
	t.Run("only first called", func(t *testing.T) {
		hOne := mockWH{
			msg:     "Handler one handled the request",
			handled: true,
		}
		hTwo := mockWH{
			msg: "Handler two should not called because already handle by hOne",
		}
		ts := httptest.NewServer(New(hOne, hTwo))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("read response body err:%v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("got status:%v want:%v", res.StatusCode, http.StatusOK)
		}
		got := string(data)
		if got != hOne.msg {
			t.Errorf("got msg:%q want:%q", got, hOne.msg)
		}
	})
	t.Run("response from hTwo", func(t *testing.T) {
		db := new(sql.DB)
		hOne := mockWH{
			msg: "Handler one just provide the db not response to the request",
			db:  db,
		}
		hTwo := mockWH{
			msg:            "Handler two response to the request",
			handled:        true,
			checkDbFromCtx: true,
		}
		hThree := mockWH{
			msg:     "Handler three should not called because already handle by hOne",
			handled: true,
		}
		ts := httptest.NewServer(New(hOne, hTwo, hThree))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("read response body err:%v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("got status:%v want:%v", res.StatusCode, http.StatusOK)
		}
		got := string(data)
		if got != hTwo.msg {
			t.Errorf("got msg:%q want:%q", got, hTwo.msg)
		}
	})
	t.Run("not handled, last handler response write by response", func(t *testing.T) {
		db := new(sql.DB)
		hOne := mockWH{
			msg: "Handler one just provide the db not response to the request",
			db:  db,
		}
		hTwo := mockWH{
			msg:            "Handler two response to the request",
			checkDbFromCtx: true,
		}
		hThree := mockWH{
			msg: "Handler three should not called because already handle by hOne",
		}
		ts := httptest.NewServer(New(hOne, hTwo, hThree))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("read response body err:%v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("got status:%v want:%v", res.StatusCode, http.StatusOK)
		}
		got := Response{}
		if err = json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.Data != hThree.msg {
			t.Errorf("got msg:%q want:%q", got.Data, hThree.msg)
		}
	})
}

type errWH struct {
	code int
	err  error
}

func (e errWH) Handle(w http.ResponseWriter, r *http.Request) Response {
	res := Response{}
	res.Error(e.err, e.code)
	return res
}

func TestReponseError(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		wh := errWH{
			code: http.StatusInternalServerError,
			err:  errors.New("test http error"),
		}
		ts := httptest.NewServer(New(wh))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != wh.code {
			t.Errorf("got status code:%v want:%v", res.StatusCode, wh.code)
		}
	})
	t.Run("app error response OK (200)", func(t *testing.T) {
		wh := errWH{
			code: http.StatusOK,
			err:  errors.New("test app error"),
		}
		ts := httptest.NewServer(New(wh))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != wh.code {
			t.Errorf("got status code:%v want:%v", res.StatusCode, wh.code)
		}
	})
}

func setGlogFlag() {
	if testing.Verbose() {
		dg := true
		debug = &dg
		flag.Set("alsologtostderr", "true")
	}
}
