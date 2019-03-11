package libpassword

import (
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	PasswordHashMatches(password, hash string) bool
}

func NewPasswordHasher() PasswordHasher {
	return passwordHasher{}
}

type passwordHasher struct {
}

func (passwordHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", status.Errorf(codes.Internal, "hash password failed: %v", err.Error())
	}
	return string(hashedPassword), nil
}

func (passwordHasher) PasswordHashMatches(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
