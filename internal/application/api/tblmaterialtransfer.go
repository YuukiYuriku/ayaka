package api

import (
	// "errors"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblmaterialtransfer"

	// "gitlab.com/ayaka/internal/domain/tbldirectpurchasercv"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblMaterialTransferApi interface {
	Fetch(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblMaterialTransferHandler struct {
	Service   service.TblMaterialTransferService   `inject:"tblMaterialTransferService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblMaterialTransferHandler) Fetch(c *fiber.Ctx) error {
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
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input material transfer")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input material transfer")
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
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch material transfer: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all material transfers")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblMaterialTransferHandler) Create(c *fiber.Ctx) error {
	var req *tblmaterialtransfer.Create
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create material transfer: %s", err.Error()))
		return err
	}

	if len(req.Details) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail material transfer is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create material transfer", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create material transfer: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create material transfer", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName)
	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create material transfer: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create material transfer", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create material transfer %s", result.DocNo))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblMaterialTransferHandler) Update(c *fiber.Ctx) error {
	var req *tblmaterialtransfer.Read

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update material transfer: %s", err.Error()))
		return err
	}
	req.DocNo = strings.ReplaceAll(code, "-", "/")

	if len(req.Details) == 0 {
		go h.Log.LogUserInfo(user.UserCode, "WARN", "Detail material transfer is empty")
		newErr := "details:Detail must contain atleast 1 data"
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create material transfer", newErr))
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update material transfer: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update material transfer", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode)
	if err != nil {
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data material transfer %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		if errors.Is(err, customerrors.ErrInvalidInput) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data material transfer %s", req.DocNo))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update material transfer: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data material transfer %s", req.DocNo))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
