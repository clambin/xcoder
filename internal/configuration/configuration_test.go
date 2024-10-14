package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	cfg, err := GetConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, "hevc-high", cfg.ProfileName)

	*videoProfile = "invalid"
	_, err = GetConfiguration()
	assert.Error(t, err)
}
