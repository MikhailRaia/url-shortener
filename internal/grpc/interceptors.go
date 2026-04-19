package grpc

import (
	"context"

	"github.com/MikhailRaia/url-shortener/internal/auth"
	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor manages user authentication for gRPC requests.
type AuthInterceptor struct {
	jwtService *auth.JWTService
}

// NewAuthInterceptor creates a new AuthInterceptor.
func NewAuthInterceptor(jwtService *auth.JWTService) *AuthInterceptor {
	return &AuthInterceptor{
		jwtService: jwtService,
	}
}

// Unary returns a unary server interceptor for authentication.
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var userID string
		var token string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				token = values[0]
				claims, err := i.jwtService.ValidateToken(token)
				if err == nil {
					userID = claims.UserID
				}
			}
		}

		if userID == "" {
			newUserID, err := generator.GenerateID(16)
			if err != nil {
				log.Error().Err(err).Msg("gRPC: Failed to generate user ID")
			} else {
				userID = newUserID
				// В gRPC мы не можем легко прокинуть куку назад через interceptor без изменения заголовков ответа,
				// но мы можем передать ID в контекст для текущего запроса.
				// Для полноценной работы клиент должен будет получить токен через какой-то механизм (например, в метаданных ответа).
			}
		}

		newCtx := context.WithValue(ctx, middleware.UserIDKey, userID)

		// Добавляем токен в метаданные ответа, если он был создан заново или обновлен
		if token == "" && userID != "" {
			newToken, err := i.jwtService.GenerateToken(userID)
			if err == nil {
				_ = grpc.SetHeader(ctx, metadata.Pairs("authorization", newToken))
			}
		}

		return handler(newCtx, req)
	}
}
