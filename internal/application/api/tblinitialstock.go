package api

import (
	// "errors"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblinitialstock"

	// "gitlab.com/ayaka/internal/domain/tblinitialstock"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblInitStockApi interface {
	Fetch(c *fiber.Ctx) error
	Detail(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblInitStockHandler struct {
	Service   service.TblInitStockService          `inject:"tblInitStockService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblInitStockHandler) Fetch(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	docNo := c.Query("search", "")
	warehouse := c.Query("warehouse", "")
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input init stock")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input init stock")
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

	result, err := h.Service.Fetch(c.Context(), docNo, warehouse, startDate, endDate, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch init stock: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all init stocks")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblInitStockHandler) Detail(c *fiber.Ctx) error {
	docNo := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	res := &tblinitialstock.Detail{
		DocNo: strings.ReplaceAll(docNo, "-", "/"),
	}

	if err := h.Validator.Validate(c.Context(), res); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Invalid request detail initial stock: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Invalid Request", err.Error()))
	}

	res, err := h.Service.Detail(c.Context(), res.DocNo)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error detail initial stock: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal server error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Get detail init stock")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, res))
}

func (h *TblInitStockHandler) Create(c *fiber.Ctx) error {
	var req *tblinitialstock.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create initial stock: %s", err.Error()))
		return err
	}

	if len(req.Detail) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail initial stock is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create initial stock", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create initial stock: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create initial stock", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create item: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create initial stock", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create initial stock %s", result.DocNo))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblInitStockHandler) Update(c *fiber.Ctx) error {
	var req *tblinitialstock.Detail

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update initial stock: %s", err.Error()))
		return err
	}
	req.DocNo = strings.ReplaceAll(code, "-", "/")

	if len(req.Detail) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail initial stock is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create initial stock", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update initial stock: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update Initial Stock", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data initial stock %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		if errors.Is(err, customerrors.ErrInvalidInput) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data initial stock %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update initial stock: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data initial stock %s", req.DocNo))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
