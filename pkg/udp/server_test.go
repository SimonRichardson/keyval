package udp

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"

	keyvalNet "github.com/SimonRichardson/keyval/pkg/net"
	"github.com/SimonRichardson/keyval/pkg/store/mocks"
	"github.com/go-kit/kit/log"
	"github.com/golang/mock/gomock"
)

func TestAPISelect(t *testing.T) {
	t.Parallel()

	t.Run("select", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9011

		// Setup server
		server := NewServer(store, log.NewNopLogger())
		listener, _ := setupServer(server, port)
		defer listener.Close()

		// Setup a client
		client := setupClient(port)
		defer client.Close()

		key := buildKey([]byte("abc"))
		value := []byte("def")

		store.EXPECT().Get(key).Return(value, true)

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(keyvalNet.Query{
			Method: keyvalNet.Select,
			Key:    key,
		}); err != nil {
			t.Fatal(err)
		}

		client.Write(buf.Bytes())

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			var buf [512]byte
			if _, err := client.Read(buf[0:]); err != nil {
				t.Fatal(err)
			}

			var res keyvalNet.Result
			dec := gob.NewDecoder(bytes.NewBuffer(buf[:]))
			if err := dec.Decode(&res); err != nil {
				t.Fatal(err)
			}

			if expected, actual := keyvalNet.OK, res.Status; expected != actual {
				t.Errorf("expected: %v, actual: %v", expected, actual)
			}
			if expected, actual := value, res.Value; !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected: %v, actual: %v", expected, actual)
			}
		}()

		wg.Wait()
	})
}

func TestAPIInsert(t *testing.T) {
	t.Parallel()

	t.Run("insert", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9012

		// Setup server
		server := NewServer(store, log.NewNopLogger())
		listener, _ := setupServer(server, port)
		defer listener.Close()

		// Setup a client
		client := setupClient(port)
		defer client.Close()

		key := buildKey([]byte("abc"))
		value := []byte("def")

		store.EXPECT().Set(key, value).Return(true)

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(keyvalNet.Query{
			Method: keyvalNet.Insert,
			Key:    key,
			Value:  value,
		}); err != nil {
			t.Fatal(err)
		}

		client.Write(buf.Bytes())

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			var buf [512]byte
			if _, err := client.Read(buf[0:]); err != nil {
				t.Fatal(err)
			}

			var res keyvalNet.Result
			dec := gob.NewDecoder(bytes.NewBuffer(buf[:]))
			if err := dec.Decode(&res); err != nil {
				t.Fatal(err)
			}

			if expected, actual := keyvalNet.Created, res.Status; expected != actual {
				t.Errorf("expected: %v, actual: %v", expected, actual)
			}
		}()

		wg.Wait()
	})
}

func TestAPIDelete(t *testing.T) {
	t.Parallel()

	t.Run("delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mocks.NewMockStore(ctrl)

		port := 9013

		// Setup server
		server := NewServer(store, log.NewNopLogger())
		listener, _ := setupServer(server, port)
		defer listener.Close()

		// Setup a client
		client := setupClient(port)
		defer client.Close()

		key := buildKey([]byte("abc"))

		store.EXPECT().Delete(key).Return(true)

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(keyvalNet.Query{
			Method: keyvalNet.Delete,
			Key:    key,
		}); err != nil {
			t.Fatal(err)
		}

		client.Write(buf.Bytes())

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			var buf [512]byte
			if _, err := client.Read(buf[0:]); err != nil {
				t.Fatal(err)
			}

			var res keyvalNet.Result
			dec := gob.NewDecoder(bytes.NewBuffer(buf[:]))
			if err := dec.Decode(&res); err != nil {
				t.Fatal(err)
			}

			if expected, actual := keyvalNet.OK, res.Status; expected != actual {
				t.Errorf("expected: %v, actual: %v", expected, actual)
			}
		}()

		wg.Wait()
	})
}

func setupServer(server *Server, port int) (*net.UDPConn, *net.UDPAddr) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}
	apiListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	go server.Serve(apiListener)
	return apiListener, udpAddr
}

func setupClient(port int) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}
	return conn
}

func buildKey(a []byte) string {
	v := base64.RawURLEncoding.EncodeToString(a)
	if v == "" {
		v = "empty"
	}
	return v
}
