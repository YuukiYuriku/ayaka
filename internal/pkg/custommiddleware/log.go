package custommiddleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/identity"
	"gitlab.com/ayaka/internal/pkg/formatter"
)

func Log(codeMap map[error]formatter.Status, statusMap map[error]int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		req := c.Request()
		resp := c.Response()
		reqBody := c.Body()
		reqHeader := req.Header.Header()
		traceID := c.Locals("requestid").(string)

		if len(traceID) < 1 {
			traceID = identity.NewID().String()
		}

		// Set context value
		c.Locals("traceId", traceID)
		c.Locals("srcIP", string(c.IP()))
		c.Locals("port", c.Port())
		c.Locals("path", c.Path())

		var err error
		if _err := c.Next(); _err != nil {
			err = _err
		}

		if c.Path() == "/ping" || c.Path() == "/ready" {
			return nil
		}

		statusCode := formatter.Success
		httpStatus := resp.StatusCode()
		if err != nil {
			httpStatus = gethttpstatus(err, statusMap)
			statusCode = getcode(err, codeMap)
		}

		logMsg := golog.LogModel{
			Header:       reqHeader,
			Request:      reqBody,
			HttpStatus:   uint64(httpStatus),
			StatusCode:   statusCode.String(),
			Response:     string(resp.Body()),
			ResponseTime: time.Since(startTime),
			Error:        err,
		}
		golog.TDR(c.Context(), logMsg)

		return err
	}
}
