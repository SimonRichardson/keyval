package tcp

import (
	"encoding/gob"
	"errors"
	"io"
)

// Query represents an encoding type for the tcp handler
type Query struct {
	Method string
	Key    string
	Value  []byte
}

// Result represents the final result of the tcp handler
type Result struct {
	Status   Status
	Value    []byte
	Duration string
}

// QueryParams defines all the dimensions of a query.
type QueryParams struct {
	Key string
}

// DecodeFrom populates a QueryParams from a URL.
func (qp *QueryParams) DecodeFrom(q Query, rb queryBehavior) error {
	// Required depending on the query behavior
	if rb == queryRequired {
		if qp.Key = q.Key; qp.Key == "" {
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
func (qr *SelectQueryResult) EncodeTo(w io.Writer) {
	enc := gob.NewEncoder(w)
	enc.Encode(Result{
		Status:   OK,
		Value:    qr.Value,
		Duration: qr.Duration,
	})
}

// InsertQueryResult contains statistics about the query.
type InsertQueryResult struct {
	Params   QueryParams
	Duration string
	Created  bool
}

// EncodeTo encodes the InsertQueryResult to the HTTP response writer.
func (qr *InsertQueryResult) EncodeTo(w io.Writer) {
	status := OK
	if qr.Created {
		status = Created
	}

	enc := gob.NewEncoder(w)
	enc.Encode(Result{
		Status:   status,
		Value:    []byte{},
		Duration: qr.Duration,
	})
}

// DeleteQueryResult contains statistics about the query.
type DeleteQueryResult struct {
	Params   QueryParams
	Duration string
}

// EncodeTo encodes the DeleteQueryResult to the HTTP response writer.
func (qr *DeleteQueryResult) EncodeTo(w io.Writer) {
	enc := gob.NewEncoder(w)
	enc.Encode(Result{
		Status:   OK,
		Value:    []byte{},
		Duration: qr.Duration,
	})
}

type queryBehavior int

const (
	queryRequired queryBehavior = iota
	queryOptional
)
