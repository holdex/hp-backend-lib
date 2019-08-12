package libauth

import (
	"context"
	"strings"
	"time"

	"bitbucket.org/holdex/hp-backend-lib/ctx"
	"bitbucket.org/holdex/hp-backend-lib/err"
	"bitbucket.org/holdex/hp-backend-lib/strings"

	"github.com/coreos/go-oidc"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/square/go-jose.v2/jwt"
)

type JWTValidator func(ctx context.Context) (context.Context, error)

func MakeJWTValidator(jwkSet oidc.KeySet, issuer string, audiences ...string) JWTValidator {
	return func(ctx context.Context) (context.Context, error) {
		token, err := extractAuth(ctx)
		if err != nil {
			return ctx, err
		}

		payload, err := jwkSet.VerifySignature(ctx, token)
		if err != nil {
			return ctx, liberr.NewUnauthenticated(err.Error())
		}

		claims := jwt.Claims{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return ctx, liberr.NewUnauthenticated(err.Error())
		}

		// Validate iss, exp, nbf, aud
		if err := claims.ValidateWithLeeway(jwt.Expected{
			Issuer:   issuer,
			Time:     time.Now(),
			Audience: audiences,
		}, 0); err != nil {
			return ctx, liberr.NewUnauthenticated("validation failed: %s", err)
		}

		return libctx.WithAuthSubject(ctx, claims.Subject), nil
	}
}

func extractAuth(ctx context.Context) (string, error) {
	val := libctx.GetVal(ctx, "authorization")
	if val == "" {
		return "", liberr.NewUnauthenticated("Request unauthenticated with authorization value")

	}
	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", liberr.NewUnauthenticated("Bad authorization value")
	}
	if strings.ToLower(splits[0]) != "bearer" {
		return "", liberr.NewUnauthenticated("Request unauthenticated with bearer")
	}
	return splits[1], nil
}

type AdminValidator func(context.Context) error

func MakeAdminValidator(adminKey string) AdminValidator {
	return func(ctx context.Context) error {
		token, err := extractAuth(ctx)
		if err != nil {
			return err
		}

		if libstrings.IsEmpty(adminKey) {
			return liberr.NewNotAuthorized("admin middleware not configured")
		}

		if adminKey != token {
			return liberr.NewUnauthenticated("admin not authorized")
		}

		return nil
	}
}
