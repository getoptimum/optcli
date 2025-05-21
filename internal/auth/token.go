package auth

import (
	"fmt"
	"time"

	"github.com/getoptimum/mump2p-cli/internal/config"
	"github.com/golang-jwt/jwt/v4"
)

// TokenParser handles JWT token parsing and validation
type TokenParser struct{}

// NewTokenParser creates a new token parser
func NewTokenParser() *TokenParser {
	return &TokenParser{}
}

// ParseToken extracts claims from a JWT token without verifying signature
func (p *TokenParser) ParseToken(tokenString string) (*TokenClaims, error) {
	// parse the token without validation
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// we don't have the key, so we can't validate signature
		// we just want to read the claims
		return nil, nil
	})

	// expect signature validation to fail
	// extract claims anyway if possible
	var claims jwt.MapClaims
	if err != nil {
		if validationError, ok := err.(*jwt.ValidationError); ok {
			if validationError.Errors == jwt.ValidationErrorSignatureInvalid {
				// this is expected, extract claims anyway
				claims, ok = token.Claims.(jwt.MapClaims)
				if !ok {
					return nil, fmt.Errorf("invalid token claims format")
				}
			} else {
				return nil, fmt.Errorf("error parsing token: %v", err)
			}
		} else {
			return nil, fmt.Errorf("error parsing token: %v", err)
		}
	} else {
		_, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims format")
		}
	}

	tc := &TokenClaims{
		IsActive:          claims["is_active"] == true,
		MaxPublishPerHour: intFromClaims(claims, "max_publish_per_hour", config.DefaultMaxPublishPerHour),
		MaxPublishPerSec:  intFromClaims(claims, "max_publish_per_sec", config.DefaultMaxPublishPerSec),
		MaxMessageSize:    int64FromClaims(claims, "max_message_size", config.DefaultMaxMessageSize),
		DailyQuota:        int64FromClaims(claims, "daily_quota", config.DefaultDailyQuota),
	}

	if sub, ok := claims["sub"].(string); ok {
		tc.Subject = sub
	}
	if iat, ok := claims["iat"].(float64); ok {
		tc.IssuedAt = time.Unix(int64(iat), 0)
	}
	if exp, ok := claims["exp"].(float64); ok {
		tc.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if cid, ok := claims["client_id"].(string); ok {
		tc.ClientID = cid
	} else if sub, ok := claims["sub"].(string); ok {
		tc.ClientID = sub // fallback if client_id missing
	}
	if lsa, ok := claims["limits_set_at"].(float64); ok {
		tc.LimitsSetAt = int64(lsa)
	}

	return tc, nil
}

func intFromClaims(c jwt.MapClaims, key string, def int) int {
	if v, ok := c[key].(float64); ok {
		return int(v)
	}
	return def
}

func int64FromClaims(c jwt.MapClaims, key string, def int64) int64 {
	if v, ok := c[key].(float64); ok {
		return int64(v)
	}
	return def
}
