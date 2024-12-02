package user

// PasswordHasher is an interface for hashing and validating passwords.
//
//go:generate mockery --name=PasswordHasher --case=underscore --inpackage
type PasswordHasher interface {
	// Hash takes a plain password and returns the hashed password and the salt used for hashing.
	// The first returned string is the hashed password, and the second returned string is the salt.
	Hash(password string) (string, string, error)

	// Validate checks if the provided password matches the hashed password when combined with the salt.
	Validate(password, hashedPassword, salt string) bool
}
