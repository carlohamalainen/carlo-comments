package simple

import (
	"context"

	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

type UserService struct {
	HmacSecret string
}

func NewUserService(hmacSecret string) *UserService {
	return &UserService{
		HmacSecret: hmacSecret,
	}
}

func (us *UserService) HashPassword(password []byte) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

func (us *UserService) Authenticate(ctx context.Context, adminUserEmail, adminUserPassword, email, password string) (*conduit.User, error) {
	logger := conduit.GetLogger(ctx)

	if adminUserEmail != email {
		logger.Info("invalid credentials", "email", email)
		return nil, fmt.Errorf("invalid credentials")
	}

	err := bcrypt.CompareHashAndPassword([]byte(adminUserPassword), []byte(password))
	if err != nil {
		logger.Info("invalid credentials", "email", email)
		return nil, fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(us.HmacSecret))
	if err != nil {
		logger.Info("token construction failed", "error", err)
		return nil, fmt.Errorf("failed to generate token")
	}

	user := conduit.User{
		Email: email,
		Token: tokenString,
	}

	return &user, nil
}
