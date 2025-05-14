package custommiddleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/domain/shared/identity"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/validator"
)

func ErrorHandler(codeMap map[error]formatter.Status, statusMap map[error]int) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		message := err.Error()
		errList := make(map[string]interface{}, 0)

		// Status code defaults to 500
		httpStatus := fiber.StatusInternalServerError

		// if error is a validator.ErrorMap
		if _err, ok := err.(*validator.ErrorMap); ok {
			message, errList = makeErrorMap(_err.Error())
			err = fiber.ErrBadRequest
		}

		// Retrieve the custom status code if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			httpStatus = e.Code
		} else {
			httpStatus = gethttpstatus(err, statusMap)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Status(httpStatus)

		code := getcode(err, codeMap)

		// traceID := c.Locals("traceId").(string)

		// fmt.Println("traced id before: ", traceID)

		// if len(traceID) < 1 {
		traceID := identity.NewID().String()
		// }

		fmt.Println("traced id error_handler after: ", traceID)

		if len(errList) > 0 {
			return c.JSON(formatter.NewErrorResponseList(code, message, traceID, errList))
		}

		return c.JSON(formatter.NewErrorResponse(code, message, traceID))
	}
}

func makeErrorMap(er string) (string, map[string]interface{}) {
	err := make(map[string]interface{}, 0)
	message := ""
	errorMsg := strings.Split(er, ";")
	for _, msg := range errorMsg {
		errorList := strings.Split(msg, ":")

		message = strings.Join(errorList[1:], ":")
		if len(errorList) > 2 {
			err[errorList[0]] = strings.Join(errorList[1:], ":")
		} else {
			err[errorList[0]] = errorList[1]
		}
	}

	return message, err
}
