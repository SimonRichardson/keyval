package http

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/SimonRichardson/keyval/pkg/store"
)

// These are the paths we're interested for the queries
const (
	APIPathSelect = "/"
	APIPathInsert = "/"
	APIPathDelete = "/"
)

// API serves the api for the underlying key/value store
type API struct {
	store store.Store
}

// NewAPI creates a API with the correct dependencies
func NewAPI(store store.Store) *API {
	return &API{
		store: store,
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	iw := &interceptingWriter{http.StatusOK, w}
	w = iw

	method, path := r.Method, r.URL.Path
	switch {
	case method == "GET" && path == APIPathSelect:
		a.handleSelect(w, r)
	case method == "PUT" && path == APIPathInsert:
		a.handleInsert(w, r)
	case method == "DELETE" && path == APIPathDelete:
		a.handleDelete(w, r)
	}
}

func (a *API) handleSelect(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	defer r.Body.Close()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, ok := a.store.Get(qp.Key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	qr := SelectQueryResult{Params: qp}
	qr.Value = value

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (a *API) handleInsert(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	qr := InsertQueryResult{Params: qp}
	qr.Created = a.store.Set(qp.Key, value)

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (a *API) handleDelete(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	defer r.Body.Close()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok := a.store.Delete(qp.Key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	qr := DeleteQueryResult{Params: qp}

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}
