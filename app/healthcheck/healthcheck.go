package healthcheck

import "context"

type Status string

const (
	StatusOK   Status = "OK"
	StatusDown Status = "DOWNED"
)

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status Status `json:"status"`
}

type HealthCheckHandler struct{}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) Handle(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{Status: StatusOK}, nil
}
