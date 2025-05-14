package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblprovince"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblProvinceAPI interface {
	FetchProvinces(c *fiber.Ctx) error
	DetailProvince(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	GetGroupProvinces(c *fiber.Ctx) error
}

type TblProvinceHandler struct {
	Service   service.TblProvinceService           `inject:"tblProvinceService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblProvinceHandler) FetchProvinces(c *fiber.Ctx) error {
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

	result, err := h.Service.FetchProvinces(c.Context(), search, param)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	user := c.Locals("user").(*jwt.Claims)
	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all provinces")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblProvinceHandler) DetailProvince(c *fiber.Ctx) error {
	provinceCode := c.Params("search")

	result, err := h.Service.DetailProvince(c.Context(), provinceCode)

	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, "Province Not Found", ""))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	user := c.Locals("user").(*jwt.Claims)
	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Get detail Province %s", provinceCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblProvinceHandler) Create(c *fiber.Ctx) error {
	var req *tblprovince.CreateTblProvince

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Create Province", err.Error()))
	}

	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create Province %s", req.ProvCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblProvinceHandler) Update(c *fiber.Ctx) error {
	code := c.Params("id")
	var req *tblprovince.UpdateTblProvince

	if err := c.BodyParser(&req); err != nil {
		return err
	}
	req.ProvCode = code

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update Province", err.Error()))
	}

	user := c.Locals("user").(*jwt.Claims)

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update Province %s", req.ProvCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblProvinceHandler) GetGroupProvinces(c *fiber.Ctx) error {
	result, err := h.Service.GetGroupProvinces(c.Context())
	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.DataNotFound, result))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	user := c.Locals("user").(*jwt.Claims)
	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch group provinces")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
