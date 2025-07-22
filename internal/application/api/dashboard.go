package api

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	share "gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type DashboardApi interface {
	Fetch(c *fiber.Ctx) error
}

type DashboardHandler struct {
	Service   service.DashboardService             `inject:"DashboardService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *DashboardHandler) Fetch(c *fiber.Ctx) error {
	user := c.Locals("user").(*share.Claims)

	result, err := h.Service.Fetch(c.Context())

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", "Internal server error get fetch dashboard")
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, err.Error(), ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Success get fetch dashboard")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
