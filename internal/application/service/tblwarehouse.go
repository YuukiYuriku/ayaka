package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblwarehouse"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblWarehouseService interface {
	FetchWarehouses(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	DetailWarehouse(ctx context.Context, code string) (*tblwarehouse.DetailTblWarehouse, error)
	Create(ctx context.Context, data *tblwarehouse.CreateTblWarehouse, userName string) (*tblwarehouse.CreateTblWarehouse, error)
	Update(ctx context.Context, data *tblwarehouse.UpdateTblWarehouse, userCode string) (*tblwarehouse.UpdateTblWarehouse, error)
}

type TblWarehouse struct {
	TemplateRepo tblwarehouse.RepositoryTblWarehouse `inject:"tblWarehouseRepository"`
}

func (s *TblWarehouse) FetchWarehouses(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchWarehouses(ctx, search, param)
}

func (s *TblWarehouse) DetailWarehouse(ctx context.Context, code string) (*tblwarehouse.DetailTblWarehouse, error) {
	return s.TemplateRepo.DetailWarehouse(ctx, code)
}

func (s *TblWarehouse) Create(ctx context.Context, data *tblwarehouse.CreateTblWarehouse, userName string) (*tblwarehouse.CreateTblWarehouse, error) {
	// Validasi input
	if data.WhsCode == "" || data.WhsName == "" {
		fmt.Println("Error: Warehouse code and name cannot be empty")
		return nil, fmt.Errorf("warehouse code and name cannot be empty")
	}

	// Metadata
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	// Log input
	fmt.Printf("Creating warehouse: Code=%s, Name=%s\n", data.WhsCode, data.WhsName)

	// Eksekusi ke repository
	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		fmt.Printf("Error creating warehouse: %s, Code=%s\n", err.Error(), data.WhsCode)
		return nil, fmt.Errorf("failed to create warehouse: %w", err) // Menggunakan error umum
	}

	// Log sukses
	fmt.Printf("Successfully created warehouse: Code=%s, Name=%s\n", data.WhsCode, data.WhsName)

	return res, nil
}

func (s *TblWarehouse) Update(ctx context.Context, data *tblwarehouse.UpdateTblWarehouse, userCode string) (*tblwarehouse.UpdateTblWarehouse, error) {
	data.LastUpBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error updating Warehouse: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
