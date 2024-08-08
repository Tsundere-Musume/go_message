package main

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
const userIdKey = contextKey("authUserID")
