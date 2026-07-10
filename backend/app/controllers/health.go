package controllers

import (
	"github.com/revel/revel"

	applicationhealth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/health"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/transport/httpdto"
)

type Health struct {
	*revel.Controller
}

func newHealthResponse(result applicationhealth.Result) httpdto.HealthResponse {
	return httpdto.HealthResponse{Status: result.Status, Service: result.Service}
}

func (c Health) Check() revel.Result {
	service := applicationhealth.NewService()
	result := service.Check(c.Request.Context())

	return c.RenderJSON(newHealthResponse(result))
}
