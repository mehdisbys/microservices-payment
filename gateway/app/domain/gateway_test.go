package domain

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

type MockAuthenticator struct {
	response bool
}

func (m *MockAuthenticator) setResponse(res bool) {
	m.response = res
}

func (m *MockAuthenticator) Authenticate(r *http.Request) (bool, error) {
	return m.response, nil
}

func TestGateway(t *testing.T) {
	r, _ := NewRequestHandler(&http.Client{}, mux.NewRouter(), &MockAuthenticator{response: true})

	tests := []struct {
		name     string
		config   Config
		expected map[string]bool
	}{
		{
			name: "register a http handler",
			config: Config{
				Urls: []URL{
					{
						Method: "GET",
						Path:   "/test/{id}",
						HTTP: &HTTP{
							Host: "test",
						},
					},
				},
			},
			expected: map[string]bool{
				"/test/{id}": false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r.Gateway(test.config)

			_ = r.GetRouter().Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
				t, err := route.GetPathTemplate()
				if err != nil {
					return err
				}
				test.expected[t] = true
				fmt.Println(t)
				return nil
			})

			for i := range test.expected {
				if !test.expected[i] {
					t.Errorf("route %s was not registered", i)
				}
			}
		})
	}
}

// test that we proxy the request for http proxy endpoint and that it returns the response and status code
func TestSyncHandler(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`ok`))
	})

	// testingHTTPClient returns a mock client that returns the
	//  response defined in `h`
	client, _ := testingHTTPClient(h)

	r, _ := NewRequestHandler(client, mux.NewRouter(), &MockAuthenticator{response: true})

	config := Config{
		Urls: []URL{
			{
				Method: "GET",
				Path:   "/test/{id}",
				HTTP: &HTTP{
					Host: "test",
				},
			},
		},
	}
	r.Gateway(config)

	req, err := http.NewRequest("GET", "/test/3", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.Handler(r.GetRouter()).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	if rr.Body.String() != string(`ok`) {
		t.Error("did not get expected response")
	}
}

func TestSyncHandlerNotMatching(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	// testingHTTPClient returns a mock client that returns the
	//  response defined in `h`
	client, close := testingHTTPClient(h)
	defer close()

	r, _ := NewRequestHandler(client, mux.NewRouter(), &MockAuthenticator{response: true})

	config := Config{
		Urls: []URL{
			{
				Method: "GET",
				Path:   "/test/{id}",
				HTTP: &HTTP{
					Host: "test",
				},
			},
		},
	}
	r.Gateway(config)

	// Note we are doing a PUT request
	req, err := http.NewRequest("PUT", "/test/3", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.Handler(r.GetRouter()).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	if rr.Body.String() == string(`ok`) {
		t.Error("did not get expected response")
	}
}

// from : https://github.com/romanyx/api_client_testing/blob/master/client_test.go
func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}
