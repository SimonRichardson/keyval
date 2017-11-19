package tcp

import (
	"encoding/gob"
	"errors"
	"io"

	"github.com/SimonRichardson/keyval/pkg/net"
)

// QueryParams defines all the dimensions of a query.
type QueryParams struct {
	Key string
}

// DecodeFrom populates a QueryParams from a URL.
func (qp *QueryParams) DecodeFrom(q net.Query, rb queryBehavior) error {
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
	enc.Encode(net.Result{
		Status:   net.OK,
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
	status := net.OK
	if qr.Created {
		status = net.Created
	}

	enc := gob.NewEncoder(w)
	enc.Encode(net.Result{
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
	enc.Encode(net.Result{
		Status:   net.OK,
		Value:    []byte{},
		Duration: qr.Duration,
	})
}

type queryBehavior int

const (
	queryRequired queryBehavior = iota
	queryOptional
)
