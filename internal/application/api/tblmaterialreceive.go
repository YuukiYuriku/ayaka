package api

import (
	// "errors"

	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblmaterialreceive"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblMaterialReceiveApi interface {
	Fetch(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
}

type TblMaterialReceiveHandler struct {
	Service   service.TblMaterialReceiveService    `inject:"tblMaterialReceiveService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblMaterialReceiveHandler) Fetch(c *fiber.Ctx) error {
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
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input material receive")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input material receive")
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
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch material receive: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all material receives")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblMaterialReceiveHandler) Create(c *fiber.Ctx) error {
	var req *tblmaterialreceive.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create material receive: %s", err.Error()))
		return err
	}

	if len(req.Details) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail material receive is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create material receive", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create material receive: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create material receive", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create material receive: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create material receive", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create material receive %s", result.DocNo))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
