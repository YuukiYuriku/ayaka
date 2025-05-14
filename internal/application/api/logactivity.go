package api

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	share "gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblLogApi interface {
	GetLog(c *fiber.Ctx) error
}

type TblLogHandler struct {
	Service   service.TblLogService                `inject:"tblLogService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblLogHandler) GetLog(c *fiber.Ctx) error {
	user := c.Locals("user").(*share.Claims)
	code := c.Query("code")
	category := c.Query("category")

	result, err := h.Service.GetLog(c.Context(), code, category)

	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error not found get log %s in %s", code, category))
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, fmt.Sprintf("No log for %s-%s Available", category, code), ""))
		}
		if errors.Is(err, customerrors.ErrKeyNotFound) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error not found get log %s in %s", code, category))
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, fmt.Sprintf("No log for %s-%s Available", category, code), ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error get log %s in %s", code, category))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, err.Error(), ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Success get log %s in %s", code, category))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
