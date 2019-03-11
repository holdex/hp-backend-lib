package libauth

import (
	"context"
	"strings"
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

type Claims struct {
	jwt.Claims
	Scopes []string `json:"scp"`
}

type ValidatorFunc func(context.Context) (context.Context, error)

func MakeAuthFuncValidator(jwkSet oidc.KeySet, issuer string) ValidatorFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := extractAuth(ctx, "bearer")
		if err != nil {
			return ctx, err
		}

		payload, err := jwkSet.VerifySignature(ctx, token)
		if err != nil {
			return ctx, status.Error(codes.PermissionDenied, err.Error())
		}

		claims := Claims{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return ctx, status.Error(codes.PermissionDenied, err.Error())
		}

		// Validate iss, exp, nbf
		if err := claims.ValidateWithLeeway(jwt.Expected{Issuer: issuer, Time: time.Now()}, 0); err != nil {
			return ctx, status.Errorf(codes.PermissionDenied, "validation failed: %s", err.Error())
		}

		return libctx.WithAuthSubject(ctx, claims.Subject), nil
	}
}

func extractAuth(ctx context.Context, expectedScheme string) (string, error) {
	val := libctx.GetVal(ctx, "authorization")
	if val == "" {
		return "", status.Error(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)

	}
	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", status.Error(codes.Unauthenticated, "Bad authorization string")
	}
	if strings.ToLower(splits[0]) != strings.ToLower(expectedScheme) {
		return "", status.Error(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	return splits[1], nil
}

type AdminValidatorFunc func(context.Context) error

func MakeAdminValidatorFunc(adminKey string) AdminValidatorFunc {
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
