package libctx

import (
	"context"
	"net/http"

	"github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"

	"github.com/holdex/hp-backend-lib/strings"
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
		ctx = WithSessionID(ctx, r.Header.Get("X-Session-ID"))
		ctx = WithUserID(ctx, r.Header.Get("X-User-ID"))
		ctx = WithGAnalyticsCID(ctx, r.Header.Get("GA-CID"))
		ctx = WithHoldexTeamMember(ctx, r.Header.Get("X-Holdex-Team-Member"))
		ctx = WithHttpReferer(ctx, r.Header.Get("X-HTTP-Referer"))
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

// Google Analytics cid
func GetGAnalyticsCID(ctx context.Context) string { return GetVal(ctx, "ga_cid") }
func WithGAnalyticsCID(ctx context.Context, userID string) context.Context {
	return setVal(ctx, "ga_cid", userID)
}

// Deprecated
func GetAuthSubject(ctx context.Context) (s string) { return GetVal(ctx, "auth_subject") }

// Deprecated
func WithAuthSubject(ctx context.Context, sub string) context.Context {
	return setVal(ctx, "auth_subject", sub)
}

func GetUserID(ctx context.Context) string { return GetVal(ctx, "user_id") }
func WithUserID(ctx context.Context, userID string) context.Context {
	return setVal(ctx, "user_id", userID)
}

func GetSessionID(ctx context.Context) string { return GetVal(ctx, "session_id") }
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return setVal(ctx, "session_id", sessionID)
}

func GetHttpReferer(ctx context.Context) string { return GetVal(ctx, "http_referer") }
func WithHttpReferer(ctx context.Context, httpReferer string) context.Context {
	return setVal(ctx, "http_referer", httpReferer)
}

func IsHoldexTeamMember(ctx context.Context) bool {
	return GetVal(ctx, "holdex_team_member") == "1"
}
func WithHoldexTeamMember(ctx context.Context, holdexTeamMember string) context.Context {
	return setVal(ctx, "holdex_team_member", holdexTeamMember)
}

func WithAuthorization(ctx context.Context, authorization string) context.Context {
	return setVal(ctx, "authorization", authorization)
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
