package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/heetch/MehdiSouilhed-technical-test/common"
)

type RequestHandler struct {
	userTokens map[string]string
}

type UserCheckAuthRequest struct {
	UserID string `json:"userID"`
	Token  string `json:"token"`
}

const (
	logTraceID = "traceID"
)

func NewRequestHandler(tokens map[string]string) RequestHandler {
	return RequestHandler{
		userTokens: tokens,
	}
}

// Authenticate checks that the user and token are valid
func (s *RequestHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	traceID := common.ExtractTraceIDFromReq(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request := UserCheckAuthRequest{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Str(logTraceID, traceID).
		Str("user", request.UserID).
		Msg("auth request")

	err = s.checkAuth(request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Info().Str(logTraceID, traceID).
		Str("user", request.UserID).
		Msg("auth successful")

	w.WriteHeader(http.StatusOK)
}

func (s *RequestHandler) checkAuth(r UserCheckAuthRequest) error {
	t, ok := s.userTokens[r.UserID]

	if !ok {
		return errors.New("user not found")
	}

	if t != r.Token {
		return errors.New("invalid token")
	}

	return nil
}
