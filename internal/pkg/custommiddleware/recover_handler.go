package custommiddleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/runsystemid/golog"
)

func RecoverHandler(c *fiber.Ctx, e interface{}) {
	err, ok := e.(error)
	if !ok {
		err = fmt.Errorf("%v", e)
	}

	golog.Error(c.Context(), err.Error(), err)
}
