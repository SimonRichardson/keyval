package udp

import (
	"bytes"
	"encoding/gob"
	"io"
	"net"
	"time"

	keyvalNet "github.com/SimonRichardson/keyval/pkg/net"
	"github.com/SimonRichardson/keyval/pkg/store"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type client struct {
	Addr  *net.UDPAddr
	Query keyvalNet.Query
}

type Server struct {
	store   store.Store
	clients chan client
	stop    chan chan struct{}
	logger  log.Logger
}

// NewServer creates a Server with the correct dependencies
func NewServer(store store.Store, logger log.Logger) *Server {
	return &Server{
		store:   store,
		clients: make(chan client, 100),
		stop:    make(chan chan struct{}),
		logger:  logger,
	}
}

// Serve the listener for the server
func (s *Server) Serve(conn *net.UDPConn) error {
	go s.handleClients(conn)
	go s.handleRequests(conn)

	for {
		select {
		case q := <-s.stop:
			close(q)
			return nil
		}
	}
}

// Stop the server
func (s *Server) Stop() {
	q := make(chan struct{})
	s.stop <- q
	<-q
}

func (s *Server) handleClients(conn *net.UDPConn) {
	for {
		select {
		case client := <-s.clients:
			addr := client.Addr
			query := client.Query

			var res bytes.Buffer
			switch query.Method {
			case keyvalNet.Select:
				s.handleSelect(&res, query)
			case keyvalNet.Insert:
				s.handleInsert(&res, query)
			case keyvalNet.Delete:
				s.handleDelete(&res, query)
			default:
				// send error
				write(conn, keyvalNet.NotFound)
			}

			if _, err := conn.WriteToUDP(res.Bytes(), addr); err != nil {
				level.Warn(s.logger).Log("err", err)
				return
			}
		}
	}
}

func (s *Server) handleRequests(conn *net.UDPConn) {
	for {
		var buf [1024]byte

		n, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			continue
		}

		dec := gob.NewDecoder(bytes.NewBuffer(buf[0:n]))
		var query keyvalNet.Query
		if err := dec.Decode(&query); err != nil {
			var res bytes.Buffer
			write(&res, keyvalNet.ServerError)
			if _, err := conn.WriteToUDP(res.Bytes(), addr); err != nil {
				level.Warn(s.logger).Log("err", err)
				continue
			}
		}

		go func() {
			s.clients <- client{
				Addr:  addr,
				Query: query,
			}
		}()
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
