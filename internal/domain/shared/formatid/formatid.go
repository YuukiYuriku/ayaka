package formatid

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/logactivity"
)

type GenerateIDModel struct {
	LastId string `db:"LastId"`
}

type GenerateIDHandler struct {
	DB *repository.Sqlx `inject:"database"`
}

func FormatId(id string, prefixParts string) (string, error) {
	splitId := strings.Split(id, "-")

	lastNumericPart := splitId[len(splitId)-1]
	num, err := strconv.Atoi(lastNumericPart)
	if err != nil {
		return "", fmt.Errorf("failed to parse numeric part: %v", err)
	}

	num++

	newId := fmt.Sprintf("%s-%05d", prefixParts, num)

	return newId, nil
}

func (s *GenerateIDHandler) GenerateID(ctx context.Context, category string) (string, error) {
	doc, err := logactivity.DocNumberOf(category)
	if err != nil {
		return "", err
	}

	primKey, err := logactivity.PrimaryKeyOf(category)
	if err != nil {
		return "", err
	}

	table, err := logactivity.TableOf(category)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf("SELECT %s AS LastId FROM %s ORDER BY %s DESC LIMIT 1", primKey, table, primKey)
	var lastId GenerateIDModel

	err = s.DB.GetContext(ctx, &lastId, query)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to get last ID: %w", err)
	}

	var splitId string
	if lastId.LastId != "" {
		parts := strings.Split(lastId.LastId, "/")
		if len(parts) > 0 {
			splitId = parts[0]
		}
	}

	if splitId == "" {
		splitId = "0"
	}

	num, err := strconv.Atoi(splitId)
	if err != nil {
		return "", fmt.Errorf("failed to parse numeric part: %v", err)
	}

	newId := fmt.Sprintf("%04d/R1/%s/%s", num+1, doc, time.Now().Format("01/06"))

	return newId, nil
}

// to get last detail number from a table
func (s *GenerateIDHandler) GetLastDetailNumber(ctx context.Context, category string) (int, error) {
	table, err := logactivity.TableOf(category)
	if err != nil {
		return -1, err
	}

	var total int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err = s.DB.GetContext(ctx, &total, query)
	if err != nil {
		return -1, fmt.Errorf("failed to get total DNo: %w", err)
	}

	return total, nil
}

// Generate Code
func (s *GenerateIDHandler) GenerateIDCode(ctx context.Context, tblName string) (string, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s;", tblName)
	var count int
	err := s.DB.GetContext(ctx, &count, query)
	if err != nil {
		return "", fmt.Errorf("failed to get count: %w", err)
	}

	if count == 0 {
		return "00001", nil
	}

	return fmt.Sprintf("%05d", count+1), nil
}