package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 30*time.Second, cfg.ReaperInterval)
	assert.Equal(t, 1*time.Hour, cfg.SandboxTTL)
}

func TestLoadConfig_InvalidDuration(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		envVal string
	}{
		{"bad reaper interval", "REAPER_INTERVAL", "not-a-duration"},
		{"bad sandbox ttl", "SANDBOX_TTL", "also-not-a-duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://test")
			t.Setenv(tt.envVar, tt.envVal)

			_, err := LoadConfig()
			assert.Error(t, err)
		})
	}
}
