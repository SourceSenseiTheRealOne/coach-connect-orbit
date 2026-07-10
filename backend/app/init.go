package app

import "github.com/revel/revel"

func init() {
	revel.Filters = []revel.Filter{
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
