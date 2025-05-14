package rest

import (
	"github.com/gofiber/fiber/v2"
	template "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/formatter"
)

var CodeMap = map[error]formatter.Status{
	// template
	template.ErrDataNotFound: formatter.DataNotFound,
}

var StatusMap = map[error]int{
	// template
	template.ErrDataNotFound: fiber.StatusNotFound,
}
