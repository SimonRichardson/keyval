package tcp

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/keyval/pkg/store/mocks"
	"github.com/go-kit/kit/log"
	"github.com/golang/mock/gomock"
)

func TestAPISelect(t *testing.T) {
	t.Parallel()

	t.Run("select with no valid content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9000

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a []byte) bool {
			key := buildKey(a)

			store.EXPECT().Get(key).Return(nil, false)

			resp := Request(port, Query{
				Method: "SELECT",
				Key:    key,
			})

			return resp.Status == NotFound
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("select", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9001

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a, b []byte) bool {
			key := buildKey(a)

			store.EXPECT().Get(key).Return(b, true)

			resp := Request(port, Query{
				Method: "SELECT",
				Key:    key,
			})

			if resp.Status != OK {
				return false
			}

			// We have to check here, because the types don't match []interface{} vs []byte because of
			// gob encoding.
			if len(b) == 0 && len(resp.Value) == 0 {
				return true
			}
			return reflect.DeepEqual(b, resp.Value)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAPIInsert(t *testing.T) {
	t.Parallel()

	t.Run("insert with no existing value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9002

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a, b []byte) bool {
			if len(b) == 0 {
				b = []byte{0}
			}

			key := buildKey(a)

			store.EXPECT().Set(key, b).Return(false)

			resp := Request(port, Query{
				Method: "INSERT",
				Key:    key,
				Value:  b,
			})

			return resp.Status == OK
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("insert with existing value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9003

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a, b []byte) bool {
			if len(b) == 0 {
				b = []byte{0}
			}

			key := buildKey(a)

			store.EXPECT().Set(key, b).Return(true)

			resp := Request(port, Query{
				Method: "INSERT",
				Key:    key,
				Value:  b,
			})

			return resp.Status == Created
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAPIDelete(t *testing.T) {
	t.Parallel()

	t.Run("delete with no existing value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9004

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a []byte) bool {
			key := buildKey(a)

			store.EXPECT().Delete(key).Return(false)

			resp := Request(port, Query{
				Method: "DELETE",
				Key:    key,
			})

			return resp.Status == NotFound
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9004

		server := NewServer(store, log.NewNopLogger())
		listener := setupServer(server, port)
		defer listener.Close()

		fn := func(a []byte) bool {
			key := buildKey(a)

			store.EXPECT().Delete(key).Return(true)

			resp := Request(port, Query{
				Method: "DELETE",
				Key:    key,
			})

			return resp.Status == OK
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func setupServer(server *Server, port int) net.Listener {
	apiListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}
	go server.Serve(apiListener)
	return apiListener
}

func buildKey(a []byte) string {
	v := base64.RawURLEncoding.EncodeToString(a)
	if v == "" {
		v = "empty"
	}
	return v
}

func Request(port int, q Query) Result {
	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	enc := gob.NewEncoder(conn)
	if err := enc.Encode(q); err != nil {
		panic(err)
	}

	var res Result
	dec := gob.NewDecoder(conn)
	if err := dec.Decode(&res); err != nil {
		panic(err)
	}

	return res
}
