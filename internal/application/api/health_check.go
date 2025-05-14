package api

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/runsystemid/gocache"
	"gitlab.com/ayaka/internal/adapter/repository"
)

type HealthCheckAPI interface {
	Ping(*fiber.Ctx) error
	Ready(*fiber.Ctx) error
	PrintVersion(c *fiber.Ctx) error
}

type HealthCheckHandler struct {
	Database *repository.Sqlx `inject:"database"`
	Cache    gocache.Service  `inject:"cache"`
}

func (h *HealthCheckHandler) Ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "pong"})
}

func (h *HealthCheckHandler) Ready(c *fiber.Ctx) error {
	message := make(map[string]string)

	sqlDB := h.Database.DB
	err := sqlDB.Ping()
	if err != nil {
		message["database"] = "not ready"
	} else {
		message["database"] = "ready"
	}

	err = h.Cache.Ping(c.Context())
	if err != nil {
		message["cache"] = "not ready"
	} else {
		message["cache"] = "ready"
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(message)
	}

	return c.Status(fiber.StatusOK).JSON(message)
}

func (h *HealthCheckHandler) PrintVersion(c *fiber.Ctx) error {
	data, err := os.ReadFile("version.txt")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read version",
		})
	}

	return c.JSON(fiber.Map{"version": string(data)})
}
