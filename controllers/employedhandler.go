package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/riyan/apiatex/controllers/dbaccess"
	"github.com/riyan/apiatex/controllers/webserver/webhandler"
	"github.com/riyan/apiatex/models"
)

//HandleGET handle GET request.
func (wh eHandler) handleGETEmployed(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	paths := UrlPath(r.URL, wh.pattern)

	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if len(paths) == 2 {
		id, err := url.PathUnescape(paths[1])
		// fmt.Printf("id : %s", id)
		if err != nil {
			res.Error(err, http.StatusInternalServerError)
			return res
		}

		if id != "" {
			Em := &models.Employed{Id: id}
			err := Em.Get(db)
			if err != nil {
				res.Error(err, http.StatusOK)
				return res
			}
			res.Data = Em
			return res
		}

	}
	var values = r.URL.Query()
	set, cursor, err := QueryCursor(values)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
	}
	if !set {
		flds, err := QueryFields(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		lmt, err := QueryLimit(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		srt, asc, err := QuerySort(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		fil, err := QueryFilter(values)
		if err != nil {
			res.Error(err, http.StatusBadRequest)
			return res
		}
		cursor = dbaccess.Cursor{
			Fields:     flds,
			Filters:    fil,
			OrderBy:    srt,
			Descending: asc,
			Limit:      lmt,
		}
	}

	var employeds []dbaccess.Table
	tem := &models.Employed{}
	employeds, cursor, err = dbaccess.Fetch(db, tem, cursor)
	if err != nil {
		res.Error(err, http.StatusOK)
		return res
	}
	res.Data = employeds
	res.Cursor = cursor
	return res
}
func (wh eHandler) handlePOSTEmployed(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	em := &models.Employed{}
	if err := json.Unmarshal(body, em); err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if err = em.Insert(tx); err != nil {
		res.Error(err, http.StatusOK)
		tx.Rollback()
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	res.Data = em
	return res
}
func (wh eHandler) handleDELETEEmployed(w http.ResponseWriter, r *http.Request) webhandler.Response {
	res := webhandler.Response{}
	paths := UrlPath(r.URL, wh.pattern)
	// fmt.Printf("len path:%d last:%s\n", len(paths), paths[len(paths)-1])
	last, err := url.PathUnescape(paths[1])
	// fmt.Printf("id : %s", id)
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	if last == "" || last == "employed" {
		res.Error(errors.New("specify id to delete"), http.StatusOK)
		return res
	}
	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	em := models.Employed{Id: last}
	if err = em.Delete(tx); err != nil {
		res.Error(err, http.StatusOK)
		tx.Rollback()
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	return res
}
func (wh eHandler) handlePUTEmployed(w http.ResponseWriter, r *http.Request) webhandler.Response {
	var res webhandler.Response
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		res.Error(err, http.StatusBadRequest)
		return res
	}
	id, ok := data["id"]
	if !ok {
		res.Error(errors.New("id must be set"), http.StatusOK)
		return res
	}
	db, err := webhandler.DBFromContext(r.Context())
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	tx, err := db.Begin()
	if err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	em := &models.Employed{Id: fmt.Sprintf("%s", id)}
	data, err = em.Update(tx, data)
	if err != nil {
		tx.Rollback()
		res.Error(err, http.StatusOK)
		return res
	}
	if err = tx.Commit(); err != nil {
		res.Error(err, http.StatusInternalServerError)
		return res
	}
	res.Data = data
	return res
}
