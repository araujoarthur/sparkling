// server.go sets up the chi router, wires middleware and registers all IAM routes.
package rest

import (
	"crypto/rsa"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/domain"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router          *chi.Mux
	roles           domain.RoleService
	permissions     domain.PermissionService
	rolePermissions domain.RolePermissionService
	principals      domain.PrincipalService
	principalRoles  domain.PrincipalRoleService
	publicKey       *rsa.PublicKey
}

// Server Constructor
func NewServer(
	publicKey *rsa.PublicKey,
	roles domain.RoleService,
	permissions domain.PermissionService,
	rolePermissions domain.RolePermissionService,
	principals domain.PrincipalService,
	principalRoles domain.PrincipalRoleService,
) *Server {
	s := &Server{
		router:          chi.NewRouter(),
		roles:           roles,
		permissions:     permissions,
		rolePermissions: rolePermissions,
		principals:      principals,
		principalRoles:  principalRoles,
		publicKey:       publicKey,
	}

	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)

	s.registerRoutes()

	return s
}

// Server's member functions

// ServeHTTP implements http.Handler so Server can be passed to http.ListenAndServe.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// registerRoutes registers all IAM routes on the chi router.
func (s *Server) registerRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		// health check — no auth required
		r.Get("/health", s.health)

		// all routes below require a valid service token
		r.Group(func(r chi.Router) {
			r.Post("/", s.createPrincipal)
			r.Use(token.Middleware(s.publicKey))

			r.Route("/roles", func(r chi.Router) {
				r.Get("/", s.listRoles)
				r.Post("/", s.createRole)
				r.Get("/{id}", s.getRoleByID)
				r.Put("/{id}", s.updateRole)
				r.Delete("/{id}", s.deleteRole)
				r.Get("/{id}/permissions", s.listPermissionsByRole)
				r.Post("/{id}/permissions", s.assignPermissionToRole)
				r.Delete("/{id}/permissions/{permID}", s.removePermissionFromRole)
				r.Get("/{id}/principals", s.listPrincipalsByRole)
			})

			r.Route("/permissions", func(r chi.Router) {
				r.Get("/", s.listPermissions)
				r.Post("/", s.createPermission)
				r.Get("/{id}", s.getPermissionByID)
				r.Delete("/{id}", s.deletePermission)
				r.Get("/{id}/roles", s.listRolesByPermission)
			})

			r.Route("/principals", func(r chi.Router) {
				r.Get("/", s.listPrincipals)
				r.Get("/by-external-id/{externalID}", s.getPrincipalByExternalID)
				r.Get("/{id}", s.getPrincipalByID)
				r.Delete("/{id}", s.deletePrincipal)
				r.Post("/{id}/activate", s.activatePrincipal)
				r.Post("/{id}/deactivate", s.deactivatePrincipal)
				r.Get("/{id}/roles", s.listRolesByPrincipal)
				r.Post("/{id}/roles", s.assignRoleToPrincipal)
				r.Delete("/{id}/roles/{roleID}", s.removeRoleFromPrincipal)
			})
		})
	})
}

// health responds with 200 OK if the service is live.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
