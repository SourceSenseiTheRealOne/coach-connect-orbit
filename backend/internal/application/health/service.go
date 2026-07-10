package health

import "context"

const serviceName = "coach-connect-api"

// Result is the framework-independent health state returned by the application layer.
type Result struct {
	Status  string
	Service string
}

// Service exposes liveness behavior without depending on Revel or infrastructure packages.
type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) Check(_ context.Context) Result {
	return Result{Status: "ok", Service: serviceName}
}
