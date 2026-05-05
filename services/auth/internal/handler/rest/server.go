package rest

import (
	"crypto/rsa"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/domain"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux

	sessions      domain.SessionService
	accounts      domain.AccountService
	serviceTokens domain.ServiceTokenService

	publicKey *rsa.PublicKey
}

func NewServer(
	publicKey *rsa.PublicKey,
	sessions domain.SessionService,
	accounts domain.AccountService,
	serviceTokens domain.ServiceTokenService,
) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		sessions:      sessions,
		accounts:      accounts,
		serviceTokens: serviceTokens,
		publicKey:     publicKey,
	}

	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)

	s.registerRoutes()

	return s
}

// ServeHTTP implements http.Handler so Server can be passed to http.ListenAndServe.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.health)

		r.Group(func(r chi.Router) {
			r.Use(token.Middleware(s.publicKey))

			r.Post("/register", s.register)
			r.Post("/login", s.login)
			r.Post("/refresh", s.refresh)
			r.Post("/logout", s.logout)
			r.Post("/logout-all", s.logoutAll)

			r.Route("/identities/{id}", func(r chi.Router) {
				r.Delete("/", s.deleteIdentity)
				r.Put("/password", s.changePassword)
				r.Delete("/credentials/{credentialID}", s.removeCredential)
			})

			r.Route("/service-tokens", func(r chi.Router) {
				r.Post("/", s.issueServiceToken)
				r.Post("/rotate", s.rotateServiceTokens)
			})
		})
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
