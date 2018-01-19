package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/riyan/apiatex/controllers/dbaccess"
	"github.com/riyan/apiatex/controllers/webserver/webhandler"
)

type eHandler struct {
	pattern string
}

func WebHandler(pattern string) http.Handler {
	wh := webhandler.New(eHandler{pattern: pattern})
	return wh
}
func (wh eHandler) Handle(w http.ResponseWriter, r *http.Request) webhandler.Response {
	var res webhandler.Response
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/e/")
	paths := strings.Split(path, "/")
	urls := paths[0]
	switch urls {
	case "employed":
		switch r.Method {
		case "GET":
			res = wh.handleGETEmployed(w, r)
		case "POST":
			res = wh.handlePOSTEmployed(w, r)
		case "PUT":
			res = wh.handlePUTEmployed(w, r)
		case "DELETE":
			res = wh.handleDELETEEmployed(w, r)
		default:
			res.Error(errors.New("Method not supported"), http.StatusBadRequest)
		}
	default:
		res.Error(errors.New("URL not valid"), http.StatusInternalServerError)
	}
	return res
}
func UrlPath(u *url.URL, pattern string) []string {

	urlpath := u.RawPath
	if urlpath == "" {
		urlpath = u.Path
	}
	pathpattern := strings.TrimPrefix(urlpath, pattern)
	path := strings.Split(pathpattern, "/")
	return path
}

//QueryFields
func QueryFields(query url.Values) ([]string, error) {
	qry := query.Get("fields")
	var flds []string
	if qry != "" {
		flds = strings.Split(qry, ",")

	}
	return flds, nil
}

//QueryCursor
func QueryCursor(query url.Values) (set bool, cursor dbaccess.Cursor, err error) {
	qry := query.Get("cursor")
	if qry != "" {
		cursor, err = dbaccess.Decode(qry)
		set = true
	}
	return set, cursor, err
}

//QueryLimit
func QueryLimit(query url.Values) (int, error) {
	qry := query.Get("limit")
	var lmt int
	var err error
	if qry != "" {
		lmt, err = strconv.Atoi(qry)
		if err != nil {
			return 0, err
		}
	}
	return lmt, nil
}

//QuerySort
func QuerySort(query url.Values) ([]string, bool, error) {
	qry := query.Get("sort")
	var srt []string
	descending := true
	if qry != "" {
		fieldsrt := strings.Split(qry, " ")
		srt = strings.Split(fieldsrt[0], ",")
		if len(srt) < 1 {
			return nil, false, fmt.Errorf("tidak memenuhi persyaratan")
		}
		if fieldsrt[1] != "DESC" {
			descending = false
		}
	}
	return srt, descending, nil
}

//QueryFilter
func QueryFilter(query url.Values) ([]dbaccess.Filter, error) {
	qry := query.Get("filters")

	var fil []dbaccess.Filter
	if qry != "" {

		fieldspr := strings.Split(qry, ";")

		for _, tc := range fieldspr {
			parameter := strings.Split(tc, ",")
			if len(parameter) != 3 {
				return nil, fmt.Errorf("tidak memenuhi persyaratan")
			}
			valueparameter, err := url.PathUnescape(parameter[2])
			// fmt.Printf("id : %s", id)
			if err != nil {
				return nil, fmt.Errorf("tidak memenuhi persyaratan")
			}
			b := dbaccess.Filter{
				Field: parameter[0],
				Op:    parameter[1],
				Value: valueparameter,
			}
			fil = append(fil, b)
		}
	}
	return fil, nil
}
func newTx(db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	if err != nil {
		return nil
	}
	return tx
}
