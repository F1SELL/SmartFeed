package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"SmartFeed/internal/domain"
	"SmartFeed/internal/transport/http/handlers"
	appmw "SmartFeed/internal/transport/http/middleware"

	// swagger docs
	_ "SmartFeed/docs"
)

type Handlers struct {
	Auth *handlers.AuthHandler
	User *handlers.UserHandler
	Post *handlers.PostHandler
	Feed *handlers.FeedHandler
}

func NewRouter(h Handlers, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmw.Metrics)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Handle("/metrics", appmw.PrometheusHandler())

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", h.Auth.Register)
			auth.Post("/login", h.Auth.Login)
		})

		api.Group(func(private chi.Router) {
			private.Use(appmw.Auth(jwtSecret))
			private.Get("/users/me", h.User.Me)
			private.Post("/users/follow/{id}", h.User.Follow)
			private.Post("/posts", h.Post.Create)
			private.Get("/feed", h.Feed.List)
		})

		api.Group(func(moderation chi.Router) {
			moderation.Use(appmw.Auth(jwtSecret))
			moderation.Use(appmw.RequireRoles(string(domain.RoleModerator), string(domain.RoleAdmin)))
			moderation.Get("/moderation/ping", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("moderation-ok"))
			})
		})
	})

	return r
}
