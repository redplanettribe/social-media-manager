package middlewares

type contextKey string

const (
	UserIDKey            contextKey = "userID"
	DeviceFingerprintKey contextKey = "deviceFingerprint"
)
