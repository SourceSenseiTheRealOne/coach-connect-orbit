package httpdto

// HealthResponse is the HTTP representation defined by the OpenAPI health contract.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}
