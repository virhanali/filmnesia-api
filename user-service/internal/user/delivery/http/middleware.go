package http

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/virhanali/filmnesia/user-service/internal/user/domain"
)

const (
	AuthUserIDKey      = "authUserID"
	AuthUsernameKey    = "authUsername"
	AuthUserRoleKey    = "authUserRole"
	AuthTokenClaimsKey = "authTokenClaims"
)

var (
	ErrMissingAuthHeader = errors.New("header authorization not found")
	ErrInvalidAuthHeader = errors.New("format header authorization not valid (must be 'Bearer {token}')")
	ErrTokenInvalid      = errors.New("token invalid or expired")
	ErrTokenParsing      = errors.New("failed to parse token")
)

func AuthMiddleware(jwtSecretKey string) gin.HandlerFunc {
	if jwtSecretKey == "" {
		log.Fatal("FATAL: JWT_SECRET_KEY cannot be empty for AuthMiddleware")
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrMissingAuthHeader.Error()})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidAuthHeader.Error()})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidAuthHeader.Error()})
			return
		}

		claims := &domain.AppClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecretKey), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Format token not valid"})
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrTokenInvalid.Error() + ": token expired"})
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrTokenInvalid.Error() + ": token not valid yet"})
			} else if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token signature invalid"})
			} else {
				log.Printf("Error parsing token: %v", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrTokenParsing.Error()})
			}
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrTokenInvalid.Error()})
			return
		}

		userID, parseErr := uuid.Parse(claims.Subject)
		if parseErr != nil {
			log.Printf("Error parsing UserID from subjesct token (Subject: %s): %v", claims.Subject, parseErr)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user information from token"})
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Set(AuthUsernameKey, claims.Username)
		c.Set(AuthUserRoleKey, claims.Role)
		c.Set(AuthTokenClaimsKey, claims)

		c.Next()
	}
}
