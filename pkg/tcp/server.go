package tcp

import (
	"encoding/gob"
	"io"
	"net"
	"time"

	keyvalNet "github.com/SimonRichardson/keyval/pkg/net"
	"github.com/SimonRichardson/keyval/pkg/store"
	"github.com/go-kit/kit/log"
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

	var query keyvalNet.Query
	if err := dec.Decode(&query); err != nil {
		// send error
		write(conn, keyvalNet.ServerError)
		return
	}

	switch query.Method {
	case keyvalNet.Select:
		s.handleSelect(conn, query)
	case keyvalNet.Insert:
		s.handleInsert(conn, query)
	case keyvalNet.Delete:
		s.handleDelete(conn, query)
	default:
		// send error
		write(conn, keyvalNet.NotFound)
	}
}

func (s *Server) handleSelect(w io.Writer, q keyvalNet.Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp keyvalNet.QueryParams
	if err := qp.DecodeFrom(q); err != nil {
		write(w, keyvalNet.BadRequest)
		return
	}

	value, ok := s.store.Get(qp.Key)
	if !ok {
		write(w, keyvalNet.NotFound)
		return
	}

	qr := keyvalNet.SelectQueryResult{Params: qp}
	qr.Value = value

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (s *Server) handleInsert(w io.Writer, q keyvalNet.Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp keyvalNet.QueryParams
	if err := qp.DecodeFrom(q); err != nil {
		write(w, keyvalNet.BadRequest)
		return
	}

	qr := keyvalNet.InsertQueryResult{Params: qp}
	qr.Created = s.store.Set(qp.Key, q.Value)

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (s *Server) handleDelete(w io.Writer, q keyvalNet.Query) {
	// useful metrics
	begin := time.Now()

	// Validate user input.
	var qp keyvalNet.QueryParams
	if err := qp.DecodeFrom(q); err != nil {
		write(w, keyvalNet.BadRequest)
		return
	}

	ok := s.store.Delete(qp.Key)
	if !ok {
		write(w, keyvalNet.NotFound)
		return
	}

	qr := keyvalNet.DeleteQueryResult{Params: qp}

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func write(w io.Writer, status keyvalNet.Status) {
	enc := gob.NewEncoder(w)
	if err := enc.Encode(keyvalNet.Result{
		Status: status,
		Value:  []byte{},
	}); err != nil {
		panic(err)
	}
}
