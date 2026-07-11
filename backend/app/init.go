package app

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/app/controllers"
	entdb "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent"
	feedapp "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/feed"
	identityapp "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/identity"
	clerkverifier "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/infrastructure/auth/clerk"
	entrepository "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/infrastructure/persistence/ent"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/revel/revel"

	observabilitylogging "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/observability/logging"
)

const serviceName = "coach-connect-api"

var (
	applicationLogger    = observabilitylogging.NewJSON(os.Stdout, serviceName, slog.LevelInfo)
	requestLoggingFilter = newRequestLoggingFilter(applicationLogger, newRequestID, time.Now)
)

func init() {
	configureFeed()
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

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type uuidGenerator struct{}

func (uuidGenerator) NewID(context.Context) (string, error)        { return uuid.NewString(), nil }
func (uuidGenerator) NewProfileID(context.Context) (string, error) { return uuid.NewString(), nil }

func configureFeed() {
	verifier, verifierErr := clerkverifier.NewVerifier(os.Getenv("CLERK_SECRET_KEY"))
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" || verifierErr != nil {
		controllers.ConfigureFeed(nil, verifier)
		return
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		controllers.ConfigureFeed(nil, verifier)
		return
	}
	client := entdb.NewClient(entdb.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repository := entrepository.NewRepository(client)
	identityService, err := identityapp.NewService(repository, uuidGenerator{}, systemClock{})
	if err != nil {
		controllers.ConfigureFeed(nil, verifier)
		return
	}
	service, err := feedapp.NewService(verifier, &identityService, repository)
	if err != nil {
		controllers.ConfigureFeed(nil, verifier)
		return
	}
	controllers.ConfigureFeed(&service, verifier)
}

var SecurityHeadersFilter = func(controller *revel.Controller, filterChain []revel.Filter) {
	headers := controller.Response.Out.Header()
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("Referrer-Policy", "no-referrer")

	filterChain[0](controller, filterChain[1:])
}
