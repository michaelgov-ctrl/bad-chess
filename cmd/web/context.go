package main

type contextKey string

const (
	isAuthenticatedContextKey   = contextKey("isAuthenticated")
	authenticatedSessionKeyName = "authenticatedUserId"
)
