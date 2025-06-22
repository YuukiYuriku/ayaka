package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	"gitlab.com/ayaka/internal/domain/tblvendorsector"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblVendorSectorApi interface {
	Fetch(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	GetSector(c *fiber.Ctx) error
	GetSubSector(c *fiber.Ctx) error
}

type TblVendorSectorHandler struct {
	Service   service.TblVendorSectorService       `inject:"tblVendorSectorService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblVendorSectorHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	search := c.Query("search")
	activeQuery := c.Query("active")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input vendor sector")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input vendor sector")
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

	active, err := strconv.ParseBool(activeQuery)
	if err != nil {
		active = false
	}

	result, err := h.Service.Fetch(c.Context(), search, active, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch vendor sector: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all vendor sector")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorSectorHandler) Create(c *fiber.Ctx) error {
	var req *tblvendorsector.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create vendor sector: %s", err.Error()))
		return err
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create vendor sector: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create vendor sector", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create vendor sector: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create vendor sector", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create vendor sector %s", req.SectorCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorSectorHandler) Update(c *fiber.Ctx) error {
	var req *tblvendorsector.Update

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update vendor sector: %s", err.Error()))
		return err
	}
	req.SectorCode = code

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update vendor sector: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update vendor sector", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data vendor sector %s", req.SectorCode))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update vendor sector: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data vendor sector %s", req.SectorCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorSectorHandler) GetSector(c *fiber.Ctx) error {
	// code := c.Query("code")
	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.GetSector(c.Context())

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch all get vendor sector: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all get vendor sector")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblVendorSectorHandler) GetSubSector(c *fiber.Ctx) error {
	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.GetSubSector(c.Context(), code)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch all get sub sector: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all get sub sector")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
