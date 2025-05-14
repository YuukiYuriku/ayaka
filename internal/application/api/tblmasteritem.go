package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tblmasteritem"

	// "gitlab.com/ayaka/internal/domain/tblitemcategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	"gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/pagination"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblItemApi interface {
	FetchItems(c *fiber.Ctx) error
	Detail(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type TblItemHandler struct {
	Service   service.TblItemService               `inject:"tblItemService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblItemHandler) FetchItems(c *fiber.Ctx) error {
	pageStr := c.Query("page", "")
	pageSizeStr := c.Query("page_size", "")
	search := c.Query("search", "")
	category := c.Query("category", "")
	activeQuery := c.Query("active")
	user := c.Locals("user").(*jwt.Claims)

	param := &pagination.PaginationParam{}
	if pageStr != "" && pageSizeStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page input items")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid Page", ""))
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid page size input items")
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

	result, err := h.Service.FetchItems(c.Context(), search, category, active, param)

	if err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error fetch items: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal Server Error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Fetch all items")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblItemHandler) Detail(c *fiber.Ctx) error {
	itemCode := c.Params("search")
	user := c.Locals("user").(*jwt.Claims)
	result, err := h.Service.Detail(c.Context(), itemCode)

	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Not found detail item: %s", itemCode))
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, "Item not found", ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error detail item: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Internal server error", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Get detail item %s", itemCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblItemHandler) Create(c *fiber.Ctx) error {
	var req *tblmasteritem.Create
	user := c.Locals("user").(*jwt.Claims)
	confirm := c.Query("confirm")

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse create item: %s", err.Error()))
		return err
	}

	conf, err := strconv.ParseBool(confirm)
	if err != nil {
		conf = false
	}

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate create item: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to create item", err.Error()))
	}

	result, err := h.Service.Create(c.Context(), req, user.UserName, conf)
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Internal server error create item: %s", err.Error()))
			return c.Status(fiber.StatusConflict).JSON(formatter.NewErrorResponse(formatter.DataConflict, err.Error(), ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error create item: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed to create item", ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Create item %s", req.ItemCode))

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}

func (h *TblItemHandler) Update(c *fiber.Ctx) error {
	var req *tblmasteritem.Update

	code := c.Params("code")
	user := c.Locals("user").(*jwt.Claims)
	confirm := c.Query("confirm")

	conf, err := strconv.ParseBool(confirm)
	if err != nil {
		conf = false
	}

	if err := c.BodyParser(&req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error parse update item: %s", err.Error()))
		return err
	}
	req.ItemCode = code

	if err := h.Validator.Validate(c.Context(), req); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Error validate update item : %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Update Item: ", err.Error()))
	}

	result, err := h.Service.Update(c.Context(), req, user.UserCode, conf)
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Internal server error create item: %s", err.Error()))
			return c.Status(fiber.StatusConflict).JSON(formatter.NewErrorResponse(formatter.DataConflict, err.Error(), ""))
		}
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data item %s", req.ItemCode))
			return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error update item: %s", err.Error()))
		return err
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", fmt.Sprintf("Update data item %s", req.ItemCode))

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, result))
}
