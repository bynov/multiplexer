package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/bynov/multiplexer/internal/multiplexer"
)

type getContentRequest struct {
	URLs []string `json:"urls"`
}

type getContentResponse struct {
	Content []string `json:"content"`
}

func decodeGetContentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req getContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	if len(req.URLs) == 0 || len(req.URLs) > 20 {
		return nil, InvalidURLLengthError(len(req.URLs))
	}

	return req, nil
}

func makeGetContentEndpoint(svc multiplexer.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getContentRequest)
		content, err := svc.GetAllContent(ctx, req.URLs)
		if err != nil {
			return nil, err
		}

		return getContentResponse{
			Content: content,
		}, nil
	}
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func NewGetContentEndpoint(svc multiplexer.Service) http.HandlerFunc {
	return httptransport.NewServer(
		makeGetContentEndpoint(svc),
		decodeGetContentRequest,
		encodeResponse,
		httptransport.ServerErrorEncoder(errorEncoder),
	).ServeHTTP
}
