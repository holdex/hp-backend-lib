package libgrpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/holdex/go-proto-validators"
	"github.com/rollbar/rollbar-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/holdex/hp-backend-lib/ctx"
	"github.com/holdex/hp-backend-lib/err"
	"github.com/holdex/hp-backend-lib/log"
	"github.com/holdex/hp-backend-lib/strings"
	"github.com/holdex/hp-backend-lib/validator"
)

func MakeLoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		started := time.Now()
		reqID := libctx.GetRequestID(ctx)
		liblog.Infof("X-Request-ID: %s %s req: %v", reqID, info.FullMethod, req)
		resp, err := handler(ctx, req)
		took := time.Since(started).Nanoseconds()
		if err == nil {
			liblog.Infof("X-Request-ID: %s %s res: %v took.ms=%.2f", reqID, info.FullMethod, resp, float64(took)/1e6)
		} else if errStatus, ok := status.FromError(err); ok {
			switch errStatus.Code() {
			case codes.OK,
				codes.AlreadyExists,
				codes.FailedPrecondition,
				codes.NotFound,
				codes.Canceled,
				codes.InvalidArgument,
				codes.PermissionDenied,
				codes.Unauthenticated:
				liblog.Infof("X-Request-ID: %s %s res: %v took.ms=%.2f", reqID, info.FullMethod, err.Error(), float64(took)/1e6)
			default:
				liblog.Errorf("X-Request-ID: %s %s res: %s took.ms=%.2f", reqID, info.FullMethod, err.Error(), float64(took)/1e6)
				rollbar.Errorf("X-Request-ID: %s %s res: %s took.ms=%.2f", reqID, info.FullMethod, err.Error(), float64(took)/1e6)
			}
		} else {
			liblog.Errorf("X-Request-ID: %s %s res: %s took.ms=%.2f", reqID, info.FullMethod, err.Error(), float64(took)/1e6)
		}
		return resp, err
	}
}

type Claims struct {
	jwt.Claims
	Scopes []string `json:"scp"`
}

func MakeAuthFuncValidator(adminSecret, audience, issuer string) grpc_auth.ServiceAuthFuncOverride {
	return &authValidatorSvc{
		jwkSet:      oidc.NewRemoteKeySet(context.Background(), issuer+".well-known/jwks.json"),
		adminSecret: adminSecret,
		audience:    audience,
		issuer:      issuer,
	}
}

type authValidatorSvc struct {
	//reg         grpcinfo.Registry
	jwkSet      oidc.KeySet
	adminSecret string
	audience    string
	issuer      string
}

func (s *authValidatorSvc) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	//var methodAuthRules auth.Rules
	//{
	//	if mi, err := s.reg.GetMethodInfo(fullMethodName).Method().GetExtension(auth.E_Rules); err != nil {
	//		if err != proto.ErrMissingExtension {
	//			return ctx, status.Errorf(codes.Internal, "could not parse auth rules: %s", err.Error())
	//		}
	//	} else if mi != nil {
	//		rules, ok := mi.(*auth.Rules)
	//		if !ok {
	//			return ctx, status.Error(codes.Internal, "failed to parse get auth rules")
	//		}
	//
	//		if rules.Skip {
	//			return ctx, nil
	//		}
	//
	//		methodAuthRules = *rules
	//	}
	//}

	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return ctx, err
	}

	//if methodAuthRules.Admin {
	//	if libstrings.IsEmpty(s.adminSecret) {
	//		return ctx, status.Error(codes.Unimplemented, "admin middleware not configured")
	//	}
	//	if token != s.adminSecret {
	//		return ctx, status.Error(codes.PermissionDenied, "not authorized")
	//	} else {
	//		return ctx, nil
	//	}
	//}

	payload, err := s.jwkSet.VerifySignature(ctx, token)
	if err != nil {
		return ctx, status.New(codes.PermissionDenied, err.Error()).Err()
	}

	claims := Claims{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ctx, status.Error(codes.PermissionDenied, err.Error())
	}

	// Validate iss, exp, nbf
	if err := claims.ValidateWithLeeway(jwt.Expected{Issuer: s.issuer, Time: time.Now()}, 0); err != nil {
		return ctx, status.Errorf(codes.PermissionDenied, "validation failed: %s", err.Error())
	}

	// Validate audience, if not public
	if libstrings.IsEmpty(s.audience) {
		return ctx, status.Error(codes.PermissionDenied, "audience missing from service middleware")
	} else if err := claims.Validate(jwt.Expected{Audience: []string{s.audience}}); err != nil {
		return ctx, status.Errorf(codes.PermissionDenied, "audience invalid: %v", err)
	}

	return libctx.WithAuthSubject(ctx, claims.Subject), nil
}

func ValidatorUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	request, ok := req.(validator.Validator)
	if ok {
		if err := request.Validate(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}

func ErrorUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		switch e := err.(type) {
		case *liberr.InvalidArgument, libvalidator.V:
			err = status.Error(codes.InvalidArgument, e.Error())
		case *liberr.NotAuthorized:
			err = status.Error(codes.PermissionDenied, e.Error())
		case *liberr.Unauthenticated:
			err = status.Error(codes.Unauthenticated, e.Error())
		}
	}
	return resp, err
}
