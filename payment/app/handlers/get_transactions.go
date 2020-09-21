package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/heetch/MehdiSouilhed-technical-test/common"
	"github.com/heetch/MehdiSouilhed-technical-test/payment/app/domain"
)

// GetTransactions returns all the transactions for a user
func (s *RequestHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	traceID := common.ExtractTraceIDFromReq(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("could not read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request := domain.GetTransactions{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("could not unmarshal request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	txs, err := s.db.GetAllTransactions(request)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("could not retrieve transactions")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(txs) == 0 {
		log.Error().Err(err).
			Str("userid", request.UserID).
			Str(logTraceID, traceID).Msg("no transactions found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response, err := json.Marshal(txs)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		log.Error().Err(err).Str(logTraceID, traceID).Msg("error")
		w.WriteHeader(http.StatusInternalServerError)
	}

}
