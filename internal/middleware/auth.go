package middleware

import (
	"context"
	"net/http"

	"github.com/MikhailRaia/url-shortener/internal/auth"
	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/rs/zerolog/log"
)

type contextKey string

// UserIDKey is the context key used to store authenticated user ID.
const UserIDKey contextKey = "userID"

// AuthMiddleware manages user authentication using JWT cookies.
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates an AuthMiddleware with the provided JWT service.
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// AuthenticateUser ensures a user is present, issuing a token and cookie if needed.
func (a *AuthMiddleware) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		log.Debug().Msg("AuthenticateUser middleware called")

		cookie, err := r.Cookie("auth_token")
		if err == nil {
			log.Debug().Msg("Found auth_token cookie")
			claims, err := a.jwtService.ValidateToken(cookie.Value)
			if err == nil {
				userID = claims.UserID
				log.Debug().Str("userID", userID).Msg("Valid token found")
			} else {
				log.Debug().Err(err).Msg("Invalid token, creating new user")
			}
		} else {
			log.Debug().Err(err).Msg("No auth_token cookie found")
		}

		if userID == "" {
			log.Debug().Msg("Creating new user")
			newUserID, err := generator.GenerateID(16)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate user ID")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			token, err := a.jwtService.GenerateToken(newUserID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate token")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400,
			})

			userID = newUserID
			log.Debug().Str("userID", userID).Msg("Created new user")
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		log.Debug().Str("userID", userID).Msg("Setting userID in context")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth enforces that a valid auth cookie is present.
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := a.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts the authenticated user ID from context.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
