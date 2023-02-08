package dependencies

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/deliveroo/test-sonarqube/internal/config"
)

func TestDependencies(t *testing.T) {
	t.Run("dependencies are correctly loaded", func(t *testing.T) {
		cfg, err := config.Load()
		assert.Nil(t, err)

		deps, err := Initialize(cfg)
		assert.Nil(t, err)
		assert.NotNil(t, deps)

		assert.Equal(t, config.Server{
			IdleTimeout:  60 * time.Second,
			Port:         3000,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 2 * time.Second,
		}, deps.Config.Server)

		assert.NotNil(t, deps.Config.Database)
		assert.NotEmpty(t, deps.Config.Database.URL)
	})
}
