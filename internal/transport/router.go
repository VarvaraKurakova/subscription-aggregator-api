package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/VarvaraKurakova/subscription-aggregator-api/docs"
	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/handler"
)

func NewRouter(subscriptionHandler *handler.SubscriptionHandler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", subscriptionHandler.Create)
		r.Get("/", subscriptionHandler.List)

		r.Get("/total", subscriptionHandler.GetTotal)

		r.Get("/{id}", subscriptionHandler.GetByID)
		r.Put("/{id}", subscriptionHandler.Update)
		r.Delete("/{id}", subscriptionHandler.Delete)
	})

	return r
}
