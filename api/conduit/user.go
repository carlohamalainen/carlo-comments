package conduit

import (
	"context"
)

type User struct {
	Email string
	Token string
}

type UserService interface {
	Authenticate(ctx context.Context, adminUserEmail, adminUserPassword, email, password string) (*User, error)
	HashPassword(password []byte) []byte
}
