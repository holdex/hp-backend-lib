package libctx

import (
	"context"
	"net/http"

	"bitbucket.org/holdex/hp-backend-lib/strings"

	"github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
)

func WithMetadata(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var xReqID string
		if reqID := r.Header.Get("X-Request-ID"); libstrings.IsEmpty(reqID) {
			xReqID = "x/" + uuid.NewV4().String()
		} else {
			xReqID = reqID
		}
		ctx = WithAuthorization(ctx, r.Header.Get("Authorization"))
		ctx = WithRequestID(ctx, xReqID)
		ctx = WithRemoteAddr(ctx, r.Header.Get("X-Real-IP"))
		ctx = WithAuthV2(ctx, r.Header.Get("AUTHV2"))
		ctx = WithUserAgent(ctx, r.UserAgent())
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetRequestID(ctx context.Context) string { return GetVal(ctx, "request_id") }
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return setVal(ctx, "request_id", requestID)
}

func GetRemoteAddr(ctx context.Context) string { return GetVal(ctx, "remote_addr") }
func WithRemoteAddr(ctx context.Context, remoteAddr string) context.Context {
	return setVal(ctx, "remote_addr", remoteAddr)
}

func GetUserAgent(ctx context.Context) string { return GetVal(ctx, "user_agent") }
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return setVal(ctx, "user_agent", userAgent)
}

func GetAuthSubject(ctx context.Context) (s string) { return GetVal(ctx, "auth_subject") }
func WithAuthSubject(ctx context.Context, sub string) context.Context {
	return setVal(ctx, "auth_subject", sub)
}

func WithAuthorization(ctx context.Context, authorization string) context.Context {
	return setVal(ctx, "authorization", authorization)
}

func IsAuthV2(ctx context.Context) bool {
	return GetVal(ctx, "authv2") == "TRUE"
}

func WithAuthV2(ctx context.Context, value string) context.Context {
	return setVal(ctx, "authv2", value)
}

func setVal(ctx context.Context, key, val string) context.Context {
	if libstrings.IsEmpty(key) || libstrings.IsEmpty(val) {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, key, val)
}

func GetVal(ctx context.Context, key string) string {
	val, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(val[key]) > 0 {
			return val[key][0]
		}
	}

	val, ok = metadata.FromOutgoingContext(ctx)
	if ok {
		if len(val[key]) > 0 {
			return val[key][0]
		}
	}
	return ""
}
