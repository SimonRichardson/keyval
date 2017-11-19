package tcp

import (
	"encoding/gob"
	"io"
	"net"
	"time"

	"github.com/SimonRichardson/keyval/pkg/store"
	"github.com/go-kit/kit/log"
)

// Status represents the different codes that can return from the handlers
type Status int

const (
	// OK code
	OK Status = iota
	// Created code
	Created
	// BadRequest err code
	BadRequest
	// NotFound err code
	NotFound
	// ServerError code
	ServerError
)

// Server represents a way to interact with the underlying key/val store over tcp
type Server struct {
	store  store.Store
	logger log.Logger
}

// NewServer creates a Server with the correct dependencies
func NewServer(store store.Store, logger log.Logger) *Server {
	return &Server{
		store:  store,
		logger: logger,
	}
}

// Serve the listener for the server
func (s *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	dec := gob.NewDecoder(conn)

	var query Query
	if err := dec.Decode(&query); err != nil {
		// send error
		write(conn, ServerError)
		return
	}

	switch query.Method {
	case "SELECT":
		s.handleSelect(conn, query)
	case "INSERT":
		s.handleInsert(conn, query)
	case "DELETE":
		s.handleDelete(conn, query)
	default:
		// send error
		write(conn, NotFound)
	}
}

func (s *Server) handleSelect(w io.Writer, q Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(q, queryRequired); err != nil {
		write(w, BadRequest)
		return
	}

	value, ok := s.store.Get(qp.Key)
	if !ok {
		write(w, NotFound)
		return
	}

	qr := SelectQueryResult{Params: qp}
	qr.Value = value

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (s *Server) handleInsert(w io.Writer, q Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(q, queryRequired); err != nil {
		write(w, BadRequest)
		return
	}

	qr := InsertQueryResult{Params: qp}
	qr.Created = s.store.Set(qp.Key, q.Value)

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (s *Server) handleDelete(w io.Writer, q Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(q, queryRequired); err != nil {
		write(w, BadRequest)
		return
	}

	ok := s.store.Delete(qp.Key)
	if !ok {
		write(w, NotFound)
		return
	}

	qr := DeleteQueryResult{Params: qp}

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func write(w io.Writer, status Status) {
	enc := gob.NewEncoder(w)
	if err := enc.Encode(Result{
		Status: status,
		Value:  []byte{},
	}); err != nil {
		panic(err)
	}
}
