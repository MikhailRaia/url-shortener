package middleware

import (
	"context"

	"github.com/MikhailRaia/url-shortener/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCAuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewGRPCAuthMiddleware(jwtService *auth.JWTService) *GRPCAuthMiddleware {
	return &GRPCAuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *GRPCAuthMiddleware) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return handler(ctx, req)
	}

	token := authHeader[0]
	claims, err := m.jwtService.ValidateToken(token)
	if err != nil {
		// Если токен невалидный, продолжаем без userID (как в HTTP)
		return handler(ctx, req)
	}

	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	return handler(ctx, req)
}
