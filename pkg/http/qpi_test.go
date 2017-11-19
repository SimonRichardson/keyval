package http

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/keyval/pkg/store/mocks"
	"github.com/golang/mock/gomock"
)

func TestAPISelect(t *testing.T) {
	t.Parallel()

	t.Run("select with no valid content", func(t *testing.T) {
		fn := func(a []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Get(key).Return(nil, false)

			resp, err := http.Get(path)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode == http.StatusNotFound
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("select", func(t *testing.T) {
		fn := func(a, b []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Get(key).Return(b, true)

			resp, err := http.Get(path)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			result, err := ioutil.ReadAll(resp.Body)

			return resp.StatusCode == http.StatusOK &&
				reflect.DeepEqual(b, result)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAPIInsert(t *testing.T) {
	t.Parallel()

	t.Run("insert with no existing value", func(t *testing.T) {
		fn := func(a, b []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Set(key, b).Return(false)

			resp, err := Put(path, b)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode == http.StatusOK
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("insert with existing value", func(t *testing.T) {
		fn := func(a, b []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Set(key, b).Return(true)

			resp, err := Put(path, b)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode == http.StatusCreated
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAPIDelete(t *testing.T) {
	t.Parallel()

	t.Run("delete with no existing value", func(t *testing.T) {
		fn := func(a []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Delete(key).Return(false)

			resp, err := Delete(path)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode == http.StatusNotFound
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("delete with existing value", func(t *testing.T) {
		fn := func(a []byte) bool {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewMockStore(ctrl)

			api := NewAPI(store)
			server := httptest.NewServer(api)
			defer server.Close()

			path, key := buildPath(server.URL, a)

			store.EXPECT().Delete(key).Return(true)

			resp, err := Delete(path)
			if err != nil {
				t.Error(err)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode == http.StatusOK
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func buildPath(serverURL string, a []byte) (string, string) {
	v := base64.RawURLEncoding.EncodeToString(a)
	if v == "" {
		v = "empty"
	}
	return fmt.Sprintf("%s/?key=%s", serverURL, v), v
}

func Put(url string, body []byte) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func Delete(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
