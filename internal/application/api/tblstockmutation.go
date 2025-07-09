package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"

	"gitlab.com/ayaka/internal/domain/tblstockmutationhdr"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblStockMutationApi interface {
	Fetch(c *fiber.Ctx) error
	Detail(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblStockMutationHandler struct {
	Service   service.TblStockMutationService      `inject:"tblStockMutationService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblStockMutationHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	docNo := c.Query("document_number", "")
	warehouse := c.Query("warehouse", "")
	// startDate := c.Query("start_date", "")
	// endDate := c.Query("end_date", "")
	batch := c.Query("batch", "")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input stock mutation")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input stock mutation")
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

	result, err := h.Service.Fetch(c.Context(), docNo, warehouse, batch, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch stock mutation: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all stock mutation")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblStockMutationHandler) Detail(c *fiber.Ctx) error {
	docNo := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	res := &tblstockmutationhdr.Detail{
		DocNo: strings.ReplaceAll(docNo, "-", "/"),
	}

	if err := h.Validator.Validate(c.Context(), res); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Invalid request detail stocck mutation: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Invalid Request", err.Error()))
	}

	res, err := h.Service.Detail(c.Context(), res.DocNo)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error detail stock mutation: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal server error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Get detail stock mutation")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, res))
}

func (h *TblStockMutationHandler) Create(c *fiber.Ctx) error {
	var req *tblstockmutationhdr.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create stock mutation: %s", err.Error()))
		return err
	}

	if len(req.FromArray) == 0 || len(req.ToArray) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail stock mutation is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create stock mutation", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create stock mutation: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create stock mutation", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidQuantity) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Invalid quantity: %s", err.Error()))
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Stock can not less than quantity", ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create stock mutation: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create stock mutation", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create stock mutation %s", result.DocNo))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblStockMutationHandler) Update(c *fiber.Ctx) error {
	var req *tblstockmutationhdr.Detail

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update stock mutation: %s", err.Error()))
		return err
	}
	req.DocNo = strings.ReplaceAll(code, "-", "/")

	if len(req.FromArray) == 0 || len(req.ToArray) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail stock mutation is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create stock mutation", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update stock mutation: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update stock mutation", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data stock mutation %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		if errors.Is(err, customerrors.ErrInvalidInput) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data stock mutation %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update stock mutation: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data stock mutation %s", req.DocNo))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}