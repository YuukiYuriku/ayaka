package api

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/application/service"
	"gitlab.com/ayaka/internal/domain/tbluser"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/formatter"
	share "gitlab.com/ayaka/internal/pkg/jwt"
	"gitlab.com/ayaka/internal/pkg/validator"
)

type TblUserApi interface {
	Login(*fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	SendEmailForgotPassword(c *fiber.Ctx) error
	ChangePassword(c *fiber.Ctx) error
}

type TblUserHandler struct {
	Service   service.TblUserService               `inject:"tblUserService"`
	Validator validator.Validator                  `inject:"validator"`
	Log       *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (h *TblUserHandler) Login(c *fiber.Ctx) error {
	var user *tbluser.Logintbluser

	if err := c.BodyParser(&user); err != nil {
		go h.Log.LogUserInfo(user.UserCode, "WARN", fmt.Sprintf("Internal server error login: %s", err.Error()))
		return err
	}

	token, duration, err := h.Service.Login(c.Context(), user)

	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidPassword) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid password login")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorLogin(formatter.InvalidRequest, "Invalid Password", "password"))
		}

		if errors.Is(err, customerrors.ErrInvalidUserCode) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Invalid usercode or username")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorLogin(formatter.InvalidRequest, "Invalid UserName/UserCode", "usercode"))
		}

		if errors.Is(err, customerrors.ErrFailedSaveToken) {
			go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error login: %s", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorLogin(formatter.InternalServerError, "Failed to Create token", ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Internal server error login: %s", err.Error()))
		return err
	}

	data := map[string]interface{}{
		"access_token": token,
		"duration":     duration,
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Login")

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, data))
}

func (h *TblUserHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		go h.Log.LogUserInfo("User", "WARN", "Unauthorized user logout: Header is required")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	// Check if the header starts with "Bearer "
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		go h.Log.LogUserInfo("User", "WARN", "Unauthorized user logout: Invalid authorization format")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization format",
		})
	}

	tokenId := parts[1]
	user := c.Locals("user").(*share.Claims)

	// Parse dan validasi token
	_, err := h.Service.Logout(c.Context(), tokenId)
	if err != nil {
		log.Printf("Detailed error: %+v", err)
		if errors.Is(err, customerrors.ErrDataNotFound) {
			go h.Log.LogUserInfo(user.UserCode, "WARN", "Token not found")
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, "Token Not Found", ""))
		}
		go h.Log.LogUserInfo(user.UserCode, "ERROR", fmt.Sprintf("Could not logout user: %s", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, fmt.Sprintf("Could Not Logout User: %s", err.Error()), ""))
	}

	go h.Log.LogUserInfo(user.UserCode, "INFO", "Logout")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, fmt.Sprintf("Success Logout %s", user.UserName)))
}

func (h *TblUserHandler) SendEmailForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `db:"Email" json:"email" validate:"email,incolumn=tbluser->Email" label:"Email"`
	}
	if err := c.BodyParser(&input); err != nil {
		go h.Log.LogUserInfo(input.Email, "ERROR", "Internal server error parse input email")
		return err
	}

	if err := h.Validator.Validate(c.Context(), input); err != nil {
		go h.Log.LogUserInfo(input.Email, "WARN", fmt.Sprintf("Failed to validate send email: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Send Email", err.Error()))
	}

	dataUser, err := h.Service.SendEmailForgotPassword(c.Context(), input.Email)
	if err != nil {
		if errors.Is(err, customerrors.ErrDataNotFound) {
			go h.Log.LogUserInfo(input.Email, "WARN", "Email not found")
			return c.Status(fiber.StatusNotFound).JSON(formatter.NewErrorResponse(formatter.DataNotFound, "Email Not Found", ""))
		}

		if errors.Is(err, customerrors.ErrFailedSaveToken) {
			go h.Log.LogUserInfo(input.Email, "ERROR", "Failed save token")
			return c.Status(fiber.StatusInternalServerError).JSON(formatter.NewErrorResponse(formatter.InternalServerError, "Failed save token", ""))
		}
		go h.Log.LogUserInfo("User", "ERROR", fmt.Sprintf("Internal server error send email: %s", err.Error()))
		return err
	}
	go h.Log.LogUserInfo(dataUser.UserCode, "INFO", "Send Forgot Password Email")

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, nil))
}

func (h *TblUserHandler) ChangePassword(c *fiber.Ctx) error {
	var user *tbluser.ForgotPassword
	if err := c.BodyParser(&user); err != nil {
		go h.Log.LogUserInfo("user", "ERROR", "Internal server error parse change password")
		return err
	}

	if err := h.Validator.Validate(c.Context(), user); err != nil {
		go h.Log.LogUserInfo("User", "WARN", fmt.Sprintf("Failed to validate change password: %s", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Failed to Change Password", err.Error()))
	}

	token := c.Params("token")

	if token == "" {
		go h.Log.LogUserInfo("User", "WARN", "Token required")
		return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Required Token", ""))
	}

	if err := h.Service.ChangePassword(c.Context(), token, user); err != nil {
		if errors.Is(err, customerrors.ErrInvalidToken) || errors.Is(err, customerrors.ErrInvalidClaims) || errors.Is(err, customerrors.ErrDataNotFound) {
			go h.Log.LogUserInfo("User", "WARN", "Invalid token")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid token", ""))
		}

		if errors.Is(err, customerrors.ErrInvalidUserCode) {
			go h.Log.LogUserInfo("User", "WARN", "Invalid user")
			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorResponse(formatter.InvalidRequest, "Invalid user", ""))
		}

		if errors.Is(err, customerrors.ErrInvalidPassword) {
			go h.Log.LogUserInfo("User", "WARN", "Password already used")
			newErr := "new_password:Password already used before;confirm_password:Password already used before"

			return c.Status(fiber.StatusBadRequest).JSON(formatter.NewErrorFieldResponse(formatter.InvalidRequest, "Password already used", newErr))
		}

		go h.Log.LogUserInfo("User", "ERROR", fmt.Sprintf("Internal server error change password: %s", err.Error()))
		return err
	}

	return c.Status(fiber.StatusOK).JSON(formatter.NewSuccessResponse(formatter.Success, "Success change password"))
}
