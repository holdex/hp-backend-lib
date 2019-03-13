package libauth

import (
	"context"
	"time"

	"bitbucket.org/holdex/hp-backend-lib/ctx"
	"bitbucket.org/holdex/hp-backend-lib/strings"
	"github.com/coreos/go-oidc"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/square/go-jose.v2/jwt"
)

type JWTValidator func(ctx context.Context) (context.Context, error)

func MakeJWTValidator(jwkSet oidc.KeySet, issuer string, audiences ...string) JWTValidator {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return ctx, err
		}

		payload, err := jwkSet.VerifySignature(ctx, token)
		if err != nil {
			return ctx, status.Error(codes.Unauthenticated, err.Error())
		}

		claims := jwt.Claims{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return ctx, status.Error(codes.Unauthenticated, err.Error())
		}

		// Validate iss, exp, nbf, aud
		if err := claims.ValidateWithLeeway(jwt.Expected{
			Issuer:   issuer,
			Time:     time.Now(),
			Audience: audiences,
		}, 0); err != nil {
			return ctx, status.Errorf(codes.Unauthenticated, "validation failed: %s", err.Error())
		}

		return libctx.WithAuthSubject(ctx, claims.Subject), nil
	}
}

type AdminValidator func(context.Context) error

func MakeAdminValidator(adminKey string) AdminValidator {
	return func(ctx context.Context) error {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return err
		}

		if libstrings.IsEmpty(adminKey) {
			return status.Errorf(codes.Unimplemented, "admin middleware not configured")
		}

		if adminKey != token {
			return status.Error(codes.Unauthenticated, "admin not authorized")
		}

		return nil
	}
}
