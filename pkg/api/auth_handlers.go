package api

import (
	"dinoc2/pkg/api/middleware"
)

// RegisterAuthRoutes registers authentication routes
func (r *Router) RegisterAuthRoutes(authMiddleware *middleware.AuthMiddleware) {
	// Register login and refresh routes
	r.routes["/api/auth/login"] = authMiddleware.HandleLogin
	r.routes["/api/auth/refresh"] = authMiddleware.HandleRefresh
}
