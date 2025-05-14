package custommiddleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/ayaka/internal/pkg/formatter"
)

func getcode(err error, codeMap map[error]formatter.Status) formatter.Status {
	for key, val := range codeMap {
		if errors.Is(err, key) {
			return val
		}
	}

	return formatter.InternalServerError
}

func gethttpstatus(err error, statusMap map[error]int) int {
	for key, val := range statusMap {
		if errors.Is(err, key) {
			return val
		}
	}

	return fiber.StatusInternalServerError
}
