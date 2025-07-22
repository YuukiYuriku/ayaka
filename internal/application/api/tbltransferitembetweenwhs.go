package api

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	// "gitlab.com/ayaka/internal/domain/stocksummary"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblTransferItemBetweenWhsApi interface {
	Fetch(c *fiber.Ctx) error
	GetMaterial(c *fiber.Ctx) error
}

type TblTransferItemBetweenWhsHandler struct {
	Service   service.TblTransferItemBetweenWhsService `inject:"tblTransferItemBetweenWhsService"`
	Validator validator.Validator                      `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler     `inject:"logActivity"`
}
// }

func (h *TblTransferItemBetweenWhsHandler) GetMaterial(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	batch := c.Query("batch", "")
	itemName := c.Query("item_name", "")
	warehouseFrom := c.Query("warehouse_from", "")
	warehouseTo := c.Query("warehouse_to", "")
	user := c.Locals("user").(*jwt.Claims)

	if warehouseFrom == "" {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Warehouse is empty")
		newErr := "warehouse:Warehouse required"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to get item", newErr))
	}

	if warehouseTo == "" {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Warehouse is empty")
		newErr := "warehouse:Warehouse required"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to get item", newErr))
	}

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input material transfer in warehouse")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input material transfer in warehouse")
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

	result, err := h.Service.GetMaterial(c.Context(), itemName, batch, warehouseFrom, warehouseTo, param)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch material transfer in warehouse: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all material transfer in warehouse")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblTransferItemBetweenWhsHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	docNo := c.Query("search", "")
	warehouseFrom := c.Query("warehouse_from", "")
	warehouseTo := c.Query("warehouse_to", "")
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input transfer between warehouse")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input transfer between warehouse")
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

	result, err := h.Service.Fetch(c.Context(), docNo, warehouseFrom, warehouseTo, startDate, endDate, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch transfer between warehouse: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all transfer between warehouses")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}