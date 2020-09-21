package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/heetch/MehdiSouilhed-technical-test/common"
	"github.com/heetch/MehdiSouilhed-technical-test/payment/app/domain"
)

type RequestHandler struct {
	db domain.DB
}

func NewRequestHandler(db domain.DB) *RequestHandler {
	return &RequestHandler{db: db}
}

const (
	logTraceID = "traceID"
)

// PayUser will send money from one user to another user
func (s *RequestHandler) PayUser(w http.ResponseWriter, r *http.Request) {
	traceID := common.ExtractTraceIDFromReq(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("could not read request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request := domain.Transaction{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("could not unmarshal request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Str(logTraceID, traceID).
		Interface("user", request).
		Str("message", "payment request")

	_, err = s.db.SaveTransaction(request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("payment failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
