package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/revel/revel"

	observabilitylogging "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/observability/logging"
)

const serviceName = "coach-connect-api"

var (
	applicationLogger    = observabilitylogging.NewJSON(os.Stdout, serviceName, slog.LevelInfo)
	requestLoggingFilter = newRequestLoggingFilter(applicationLogger, newRequestID, time.Now)
)

func init() {
	revel.Filters = []revel.Filter{
		requestLoggingFilter,
		revel.PanicFilter,
		revel.RouterFilter,
		revel.FilterConfiguringFilter,
		revel.ParamsFilter,
		SecurityHeadersFilter,
		revel.InterceptorFilter,
		revel.CompressFilter,
		revel.BeforeAfterFilter,
		revel.ActionInvoker,
	}
}

var SecurityHeadersFilter = func(controller *revel.Controller, filterChain []revel.Filter) {
	headers := controller.Response.Out.Header()
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("Referrer-Policy", "no-referrer")

	filterChain[0](controller, filterChain[1:])
}
