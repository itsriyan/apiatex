package webhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/riyan/apiatex/controllers/dbaccess"
)

var (
	// ErrCtxNoDB is error returned by DBFromContext when ctx does not have a *sql.DB.
	ErrCtxNoDB = errors.New("ctx does not have a db")
)
var (
	//Debug set to the true to log more detail error message.
	//Set the debug only one when the server start.
	debug *bool
)

// DebugOn  call this will make response to the http client and log record with more
// detailed information.Call multiple time on this will panic.
func DebugOn() {
	if debug != nil {
		panic("DebugOn called multiple time")
	}
	dbg := true
	debug = &dbg
}

var httpErrorMessage = map[int]string{
	http.StatusNoContent:           "204 no content",
	http.StatusBadRequest:          "400 bad request",
	http.StatusUnauthorized:        "401 unathorize",
	http.StatusForbidden:           "403 forbidden",
	http.StatusNotFound:            "404 page not found",
	http.StatusMethodNotAllowed:    "405 method not allowed",
	http.StatusRequestTimeout:      "408 request timeout",
	http.StatusInternalServerError: "500 internal server error",
	http.StatusServiceUnavailable:  "503 service unavailable",
}

func httpErrMessage(code int) string {
	if errm, ok := httpErrorMessage[code]; ok {
		return errm
	}
	return fmt.Sprintf("unknown error status code:%d", code)
}

type httpError struct {
	//err is original error
	msg  string
	code int
	err  error
}

func (h httpError) Error() string {
	if debug != nil && h.err != nil {
		return h.err.Error()
	}
	return h.msg
}

func newHttpError(err error, code int) error {
	he := httpError{
		code: code,
		msg:  httpErrMessage(code),
		err:  err,
	}
	return he
}

type WebHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) Response
}

type contextKey int

const (
	dbContextKey contextKey = iota
)

// NewContextWithDB return a new context with the *sql.DB.
func NewContextWithDB(ctx context.Context, db *sql.DB) context.Context {
	ctx = context.WithValue(ctx, dbContextKey, db)
	return ctx
}

// DBFromContext return an error ErrCtxNoDB ,if ctx does not have *sql.DB.
func DBFromContext(ctx context.Context) (*sql.DB, error) {
	db, ok := ctx.Value(dbContextKey).(*sql.DB)
	if !ok {
		return nil, ErrCtxNoDB
	}
	return db, nil
}

// Response is response from handler sent to the client.
// Response will write to the http.ResponseWriter if Handled is false.
type Response struct {
	//Handled set to true if the request laready handled by handler.
	Handled bool `json:"-"`
	//Ctx is a new context based on existing context from Request , this new context
	//will be sent to the next handler.
	Ctx context.Context `json:"-"`
	// ErrMessage will be set when Error method is called, do not set it directly.
	ErrMessage string `json:"err"`
	// Cursor is query state to get the bext data from the database.
	Cursor dbaccess.Cursor `json:"cursor"`
	//Err is wrapped error.
	err           error
	Data          interface{} `json:"data"`
	handleStart   time.Time
	handleDur     time.Duration
	responseStart time.Time
	responseDur   time.Duration
}

// Error wrap the err, so it can print stack trace when debug is on.
// IF status code is http.StatusOK (200)
func (res *Response) Error(err error, code int) {
	if err != nil {
		res.ErrMessage = err.Error()
	}
	if code == http.StatusOK {
		res.err = errors.Wrap(err, "app error")
		return
	}
	err = newHttpError(err, code)
	res.err = errors.Wrap(err, "http error")
}

func (res Response) write(w http.ResponseWriter, r *http.Request) error {
	if res.err != nil {
		if he, ok := errors.Cause(res.err).(httpError); ok {
			http.Error(w, he.msg, he.code)
			return nil
		}
	}
	return res.json(w)
}

func (res Response) json(w http.ResponseWriter) error {
	b, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(b)
	return err
}

// New allocate a new http.Handler, if the handlers is more the one new will chain the
// handles as one http.Handler, it call sequentially from the beginning (index 0) to the
// end, when handler set Handled or call Error method on Reponse it will break the sequence
// (will not call the next handler).IF there is no handlers new will panic.
func New(handlers ...WebHandler) http.Handler {
	if len(handlers) == 0 {
		panic("no webhandler provided")
	}
	wh := webHandler{handlers: handlers}
	return wh
}

type webHandler struct {
	handlers []WebHandler
}

var handlerDB *sql.DB

//RegisterDB register the dabatase use as resource for the handler.
func RegisterDB(db *sql.DB) {
	if handlerDB != nil {
		panic("db already regitered on webhandler")
	}
	handlerDB = db
}

func (wh webHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var res Response
	res.Ctx = NewContextWithDB(r.Context(), handlerDB)
	start := time.Now()
	for _, handler := range wh.handlers {
		if res.Handled || res.err != nil {
			break
		}
		if res.Ctx != nil {
			r = r.WithContext(res.Ctx)
		}
		res = handler.Handle(w, r)
	}
	res.handleStart = start
	res.responseStart = time.Now()
	res.handleDur = res.responseStart.Sub(res.handleStart)
	var err error
	if !res.Handled {
		err = res.write(w, r)
	}
	res.responseDur = time.Since(res.responseStart)
	wh.log(r, res, err)
}

func (wh webHandler) log(r *http.Request, res Response, err error) {
	if debug != nil {
		wh.logDebug(r, res, err)
		return
	}
	if err != nil {
		glog.Errorf("writeErr:%s method:%s url:%s hs:%s hd:%s rs:%s rd:%s rerr:%v", err.Error(), r.Method, r.URL,
			res.handleStart, res.handleDur, res.responseStart, res.responseDur, res.err)
	} else {
		glog.Infof("method:%s url:%s hs:%s hd:%s rs:%s rd:%s rerr:%v", r.Method, r.URL,
			res.handleStart, res.handleDur, res.responseStart, res.responseDur, res.err)
	}
}

func (wh webHandler) logDebug(r *http.Request, res Response, err error) {
	if err != nil {
		glog.Errorf("writeErr:%s method:%s url:%s hs:%s hd:%s rs:%s rd:%s rerr:%+v", err.Error(), r.Method, r.URL,
			res.handleStart, res.handleDur, res.responseStart, res.responseDur, res.err)
	} else {
		glog.Infof("method:%s url:%s hs:%s hd:%s rs:%s rd:%s rerr:%+v", r.Method, r.URL,
			res.handleStart, res.handleDur, res.responseStart, res.responseDur, res.err)
	}
}
