//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	crt "gitlab.com/DzmitryYafremenka/golang-united-school-certs"
)

func Test_DBConnection(t *testing.T) {
	t.Run("Check pool connection with ping", func(t *testing.T) {
		connString := "postgres://user:password@localhost:5432/registry"
		_, err := crt.NewDirectRegistry(connString)
		assert.NoError(t, err)
	})
}
