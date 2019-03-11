package libjwt

import "github.com/dgrijalva/jwt-go"

type Service interface {
	New(jwt.Claims) string
	Parse(c jwt.Claims, token string) error
}

func NewService(signingKey string) Service {
	return &service{[]byte(signingKey)}
}

type service struct {
	signingKey []byte
}

func (s *service) New(claims jwt.Claims) string {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, _ := jwtToken.SignedString(s.signingKey)
	return token
}

func (s *service) Parse(claims jwt.Claims, token string) error {
	_, err := jwt.ParseWithClaims(token, claims,
		func(token *jwt.Token) (interface{}, error) {
			return s.signingKey, nil
		},
	)
	return err
}
