package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/ayaka/internal/pkg/formatter"
)

func TestResponse(t *testing.T) {
	t.Run("new success response", func(t *testing.T) {
		res := formatter.NewSuccessResponse(formatter.Success, "success")

		assert.Equal(t, res.Status, "00")
	})

	t.Run("new error response", func(t *testing.T) {
		res := formatter.NewErrorResponse(formatter.InternalServerError, "unexpected", "12345")

		assert.Equal(t, res.Status, "PAKU05")
		assert.Equal(t, res.Message, "unexpected")
	})
}
