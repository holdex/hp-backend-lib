package libauth

import (
	"context"

	"github.com/holdex/hp-backend-lib/ctx"
	"github.com/holdex/hp-backend-lib/err"
	"github.com/holdex/hp-backend-lib/strings"
)

type SessionValidator func(ctx context.Context) error

func MakeSessionValidator() SessionValidator {
	return func(ctx context.Context) error {
		if libstrings.IsEmpty(libctx.GetSessionID(ctx)) ||
			libstrings.IsEmpty(libctx.GetUserID(ctx)) {
			return liberr.NewUnauthenticated("missing session")
		}
		return nil
	}
}
