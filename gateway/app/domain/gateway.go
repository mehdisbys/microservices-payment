package domain

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/heetch/MehdiSouilhed-technical-test/auth/auth/handlers"
	"github.com/heetch/MehdiSouilhed-technical-test/common"
)

const (
	logTraceID = "traceID"
)

type RequestHandler struct {
	client *http.Client
	router *mux.Router
	auth   Authenticator
}

type Authenticator interface {
	Authenticate(r *http.Request) (bool, error)
}

type Auth struct {
	client *http.Client
}

func NewAuth(client *http.Client) *Auth {
	return &Auth{client: client}
}

type Message struct {
	Body       []byte            `json:"body"`
	Parameters map[string]string `json:"parameters"`
}

func NewRequestHandler(client *http.Client, r *mux.Router, auth Authenticator) (*RequestHandler, error) {
	return &RequestHandler{
		client: client,
		router: r,
		auth:   auth,
	}, nil
}

func (s *RequestHandler) GetRouter() *mux.Router {
	return s.router
}

func (s *RequestHandler) Gateway(config Config) {
	for _, c := range config.Urls {
		host := c.HTTP.Host
		s.makeSyncHandler(c.Method, c.Path, host)
	}
}

func (s *RequestHandler) makeSyncHandler(method, path, host string) {

	log.Info().Msgf("Registering http proxy handler for [method|path|host]: [%s|%s|%s]", method, path, host)

	s.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

		if r.Method != method {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		traceID := common.ExtractTraceIDFromReq(r)

		res, err := s.proxy("http://"+host+r.URL.Path, r)
		if err != nil {
			log.Error().Err(err).Str(logTraceID, traceID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// copy headers
		for name := range res.Header {
			w.Header().Set(name, res.Header[name][0])
		}

		defer res.Body.Close()

		respBytes, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Error().Err(err).Str(logTraceID, traceID)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(res.StatusCode)

		_, err = w.Write(respBytes)
		if err != nil {
			log.Error().Err(err).Str(logTraceID, traceID)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (s *RequestHandler) proxy(proxyURL string, r *http.Request) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, proxyURL, r.Body)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return nil, err
	}

	params := r.URL.Query()
	req.URL.RawQuery = params.Encode()

	valid, err := s.auth.Authenticate(r)
	if err != nil {
		return nil, err
	}

	if !valid {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	}

	// Pass the traceID downstream
	req.Header.Add(common.TraceIDHeader, common.ExtractTraceIDFromReq(r))

	response, err := s.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return nil, err
	}

	return response, nil
}
func (a *Auth) Authenticate(r *http.Request) (bool, error) {
	authURL := "http://auth/authenticate"
	request := handlers.UserCheckAuthRequest{
		UserID: r.Header.Get("X-User-Id"),
		Token:  r.Header.Get("Authorization"),
	}

	body, err := json.Marshal(request)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewReader(body))
	if err != nil {
		log.Error().Err(err).Msg("error")
		return false, err
	}

	log.Info().Interface("request", request).Msg("authenticating request")

	req.Header.Add(common.TraceIDHeader, common.ExtractTraceIDFromReq(r))

	response, err := a.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return false, err
	}
	defer response.Body.Close()

	log.Info().Interface("request", request).
		Int("status", response.StatusCode).
		Msg("authentication result")

	return response.StatusCode == http.StatusOK, nil
}
