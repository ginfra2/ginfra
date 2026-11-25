package configctl

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"testing"
)

func TestSetSpecifics(t *testing.T) {
	t.Run("Set - Bad config provider", func(t *testing.T) {
		err := SetSpecifics(app.GetRootContext(), "borkbork", "x", "x", "x", "x",
			"x")
		assert.ErrorContains(t, err, "Could not load configuration provider")
	})
}
