package api

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblHistoryOfStockApi interface {
	Fetch(c *fiber.Ctx) error
}

type TblHistoryOfStockHandler struct {
	Service   service.TblHistoryOfStockService     `inject:"tblHistoryOfStockService"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblHistoryOfStockHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	item := c.Query("item")
	batch := c.Query("batch")
	source := c.Query("source")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input history of stock")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input history of stock")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page Size", ""))
		}

		// Validasi nilai
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}

		param = &pagination.PaginationParam{
			Page:     page,
			PageSize: pageSize,
		}
	} else {
		param = nil
	}

	result, err := h.Service.Fetch(c.Context(), item, batch, source, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch history of stock: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all history of stock")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}