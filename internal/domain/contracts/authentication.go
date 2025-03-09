package contracts

import (
	"context"
)

type AuthenticationProvider interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (bool, map[string]interface{}, error)
	RevokeToken(ctx context.Context, token string) error
}
