package grpcserver

import (
	"context"
	"strings"

	"github.com/coolycow/shortener/internal/service"
	"github.com/coolycow/shortener/internal/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataSetAuthorization — нестандартный исходящий заголовок с hex-токеном для следующих вызовов (аналог Set-Cookie).
const metadataSetAuthorization = "set-authorization"

// AuthUnaryServerInterceptor подставляет userID в контекст по правилам OptionalAuth (HTTP).
func AuthUnaryServerInterceptor(users service.UserService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		auth := ""
		if v := md.Get("authorization"); len(v) > 0 {
			auth = v[0]
		}

		userID, newToken, err := resolveGRPCUser(ctx, users, auth)
		if err != nil {
			return nil, err
		}

		if newToken != "" {
			if herr := grpc.SetHeader(ctx, metadata.Pairs(metadataSetAuthorization, newToken)); herr != nil {
				return nil, status.Errorf(codes.Internal, "set header: %v", herr)
			}
		}

		ctx = shortener.ContextWithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// resolveGRPCUser: пустой/невалидный токен → новый пользователь и новый токен; валидный, но user не найден → Unauthenticated.
func resolveGRPCUser(ctx context.Context, users service.UserService, authHeader string) (userID string, newAuthHex string, err error) {
	authHeader = strings.TrimSpace(authHeader)

	userID, decErr := users.GetUserIDFromAuthToken(authHeader)
	if decErr != nil {
		u, cerr := users.CreateUser(ctx)
		if cerr != nil {
			return "", "", status.Errorf(codes.Internal, "create user: %v", cerr)
		}

		tok, terr := users.GetCookieValueByUser(u)
		if terr != nil {
			return "", "", status.Errorf(codes.Internal, "issue token: %v", terr)
		}

		return u.ID, tok, nil
	}

	if _, gerr := users.GetUser(ctx, userID); gerr != nil {
		return "", "", status.Error(codes.Unauthenticated, "user not found")
	}

	return userID, "", nil
}
