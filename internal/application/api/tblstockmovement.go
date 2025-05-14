package api

import (
	"errors"
	"fmt"
	"strconv"

	// "strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	// "gitlab.com/ayaka/internal/domain/stockMovement"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblStockMovementApi interface {
	Fetch(c *fiber.Ctx) error
	// GetItem(c *fiber.Ctx) error
}

type TblStockMovementHandler struct {
	Service   service.TblStockMovementService      `inject:"tblStockMovementService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblStockMovementHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	warehouse := c.Query("warehouse", "[]")
	dateRangeStart := c.Query("date_start", "")
	dateRangeEnd := c.Query("date_end", "")
	docType := c.Query("doc_type", "")
	itemCatCode := c.Query("item_category", "")
	batch := c.Query("batch", "")
	itemName := c.Query("item_name", "")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input stock Movement")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input stock Movement")
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

	result, err := h.Service.Fetch(c.Context(), warehouse, dateRangeStart, dateRangeEnd, docType, itemCatCode, itemName, batch, param)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidArrayFormat) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid warehouse input stock Movement")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.InvalidRequest, customerrors.ErrInvalidArrayFormat.Error())
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch stock Movement: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all stock Movement")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
