package http

import (
	"errors"
	"net/http"
	"net/url"
)

// QueryParams defines all the dimensions of a query.
type QueryParams struct {
	Key string
}

// DecodeFrom populates a QueryParams from a URL.
func (qp *QueryParams) DecodeFrom(u *url.URL, rb queryBehavior) error {
	// Required depending on the query behavior
	if rb == queryRequired {
		if qp.Key = u.Query().Get("key"); qp.Key == "" {
			return errors.New("error reading 'key' (required) query")
		}
	}
	return nil
}

// SelectQueryResult contains statistics about the query.
type SelectQueryResult struct {
	Params   QueryParams
	Duration string
	Value    []byte
}

// EncodeTo encodes the SelectQueryResult to the HTTP response writer.
func (qr *SelectQueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderDuration, qr.Duration)
	w.Header().Set(httpHeaderKey, qr.Params.Key)

	if _, err := w.Write(qr.Value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// InsertQueryResult contains statistics about the query.
type InsertQueryResult struct {
	Params   QueryParams
	Duration string
	Created  bool
}

// EncodeTo encodes the InsertQueryResult to the HTTP response writer.
func (qr *InsertQueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderDuration, qr.Duration)
	w.Header().Set(httpHeaderKey, qr.Params.Key)

	if qr.Created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// DeleteQueryResult contains statistics about the query.
type DeleteQueryResult struct {
	Params   QueryParams
	Duration string
}

// EncodeTo encodes the DeleteQueryResult to the HTTP response writer.
func (qr *DeleteQueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderDuration, qr.Duration)
	w.Header().Set(httpHeaderKey, qr.Params.Key)
}

const (
	httpHeaderDuration = "X-Duration"
	httpHeaderKey      = "X-Key"
)

type queryBehavior int

const (
	queryRequired queryBehavior = iota
	queryOptional
)
