package client

import (
	"context"
)

// tokenAuth implements grpc.PerRPCCredentials
type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return false // 允许在非 TLS 下传（视官方服务端需求而定，通常开发环境需要设为 false）
}
