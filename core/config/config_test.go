package config

import (
	"testing"
)

func TestUsesTLS(t *testing.T) {
	tests := []struct {
		name string
		cfg  CoreConfig
		want bool
	}{
		{
			name: "https endpoint",
			cfg:  CoreConfig{PermifyURL: "https://permify.example.com"},
			want: true,
		},
		{
			name: "grpcs endpoint",
			cfg:  CoreConfig{PermifyURL: "grpcs://permify.example.com"},
			want: true,
		},
		{
			name: "client certificate pair",
			cfg:  CoreConfig{PermifyURL: "localhost:3478", CertPath: "client.crt", CertKey: "client.key"},
			want: true,
		},
		{
			name: "plain localhost",
			cfg:  CoreConfig{PermifyURL: "localhost:3478"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usesTLS(tt.cfg); got != tt.want {
				t.Fatalf("usesTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}
