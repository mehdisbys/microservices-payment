package common

import (
	"net/http"

	"github.com/satori/go.uuid"
)

const TraceIDHeader = "X-Trace-Id"

func ExtractTraceIDFromReq(r *http.Request) (traceID string) {
	traceID = r.Header.Get(TraceIDHeader)
	if traceID == "" {
		traceID = uuid.NewV4().String()
	}
	return traceID
}
