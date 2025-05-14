package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblwarehousecategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblWarehouseCategoryApi interface {
	FetchWarehouseCategory(c *fiber.Ctx) error
	DetailWarehouseCategory(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblWarehouseCategoryHandler struct {
	Service   service.TblWarehouseCategoryService  `inject:"tblWarehouseCategoryService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblWarehouseCategoryHandler) FetchWarehouseCategory(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	search := c.Query("search")

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
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

	result, err := h.Service.FetchWarehouseCategory(c.Context(), search, param)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	user := c.Locals("user").(*jwt.Claims)
	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all warehouse categories")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblWarehouseCategoryHandler) DetailWarehouseCategory(c *fiber.Ctx) error {
	whsCtCode := c.Params("code")

	result, err := h.Service.DetailWarehouseCategory(c.Context(), whsCtCode)

	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, "Warehouse Category Not Found", ""))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	user := c.Locals("user").(*jwt.Claims)
	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Get detail warehouse category %s", whsCtCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblWarehouseCategoryHandler) Create(c *fiber.Ctx) error {
	var req *tblwarehousecategory.CreateTblWarehouseCategory

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Create Warehouse Category", err.Error()))
	}

	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to Create Warehouse Category", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create warehouse category %s", req.WhsCtCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblWarehouseCategoryHandler) Update(c *fiber.Ctx) error {
	var req *tblwarehousecategory.UpdateTblWarehouseCategory

	code := c.Params("code")

	if err := c.BodyParser(&req); err != nil {
		return err
	}
	req.WhsCtCode = code

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update Warehouse Category", err.Error()))
	}

	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update warehouse category %s", req.WhsCtCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
