package models

type AuthProvider string

var (
	AuthProviderCredential AuthProvider = "credential"
	AuthProviderGoogle     AuthProvider = "google"
)
