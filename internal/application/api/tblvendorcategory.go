package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	"gitlab.com/ayaka/internal/domain/tblvendorcategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblVendorCategoryApi interface {
	Fetch(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblVendorCategoryHandler struct {
	Service   service.TblVendorCategoryService     `inject:"tblVendorCategoryService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblVendorCategoryHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	search := c.Query("search")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input vendor category")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input vendor category")
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

	result, err := h.Service.Fetch(c.Context(), search, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch vendor category: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all vendor category")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorCategoryHandler) Create(c *fiber.Ctx) error {
	var req *tblvendorcategory.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create vendor category: %s", err.Error()))
		return err
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create vendor category: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create vendor category", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create vendor category: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create vendor category", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create vendor category %s", req.VendorCategoryCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorCategoryHandler) Update(c *fiber.Ctx) error {
	var req *tblvendorcategory.Update

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update vendor category: %s", err.Error()))
		return err
	}
	req.VendorCategoryCode = code

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update vendor category: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update vendor category", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data vendor category %s", req.VendorCategoryCode))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update vendor category: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data vendor category %s", req.VendorCategoryCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
