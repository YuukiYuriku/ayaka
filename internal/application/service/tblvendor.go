package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblmastervendor"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorService interface {
	Fetch(ctx context.Context, name, cat string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, vendorCode string) (*tblmastervendor.Detail, error)
	Create(ctx context.Context, data *tblmastervendor.Create, userName string) (*tblmastervendor.Create, error)
	Update(ctx context.Context, data *tblmastervendor.Update, userCode string) (*tblmastervendor.Update, error)
	GetContact(ctx context.Context, code string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblVendor struct {
	TemplateRepo tblmastervendor.Repository  `inject:"tblVendorRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblVendor) Create(ctx context.Context, data *tblmastervendor.Create, userName string) (*tblmastervendor.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	// Generate VendorCode
	vendorCode, err := s.ID.GenerateIDCode(ctx, "tblvendorhdr")
	if err != nil {
		golog.Error(ctx, "Error generate id create initial stock: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create initial stock: " + err.Error())
	}
	data.VendorCode = vendorCode

	// setup nullable and boolean
	data.Address.SetNullIfEmpty()
	data.PostalCode.SetNullIfEmpty()
	data.Website.SetNullIfEmpty()
	data.HeadOffice.SetNullIfEmpty()
	data.Phone.SetNullIfEmpty()
	data.Mobile.SetNullIfEmpty()
	data.Email.SetNullIfEmpty()
	data.Remark.SetNullIfEmpty()

	if len(data.ContactVendor) != 0 {
		for i := range data.ContactVendor {
			data.ContactVendor[i].Active = booldatatype.FromBool(true)
			data.ContactVendor[i].Position.SetNullIfEmpty()
			data.ContactVendor[i].Type.SetNullIfEmpty()
		}
	}

	if len(data.ItemCategoryVendor) != 0 {
		var temp []tblmastervendor.ItemCategoryVendor

		for i := range data.ItemCategoryVendor {
			if data.ItemCategoryVendor[i].ItemCategoryCode != "" {
				temp = append(temp, data.ItemCategoryVendor[i])
			}
		}

		data.ItemCategoryVendor = temp

		for i := range data.ItemCategoryVendor {
			data.ItemCategoryVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	if len(data.SectorVendor) != 0 {
		for i := range data.SectorVendor {
			data.SectorVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	if len(data.RatingVendor) != 0 {
		for i := range data.RatingVendor {
			data.RatingVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	return s.TemplateRepo.Create(ctx, data)
}

func (s *TblVendor) Fetch(ctx context.Context, name, cat string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, name, cat, param)
}

func (s *TblVendor) Detail(ctx context.Context, vendorCode string) (*tblmastervendor.Detail, error) {
	return s.TemplateRepo.Detail(ctx, vendorCode)
}

func (s *TblVendor) Update(ctx context.Context, data *tblmastervendor.Update, userCode string) (*tblmastervendor.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	// setup nullable and boolean
	data.Address.SetNullIfEmpty()
	data.PostalCode.SetNullIfEmpty()
	data.Website.SetNullIfEmpty()
	data.HeadOffice.SetNullIfEmpty()
	data.Phone.SetNullIfEmpty()
	data.Mobile.SetNullIfEmpty()
	data.Email.SetNullIfEmpty()
	data.Remark.SetNullIfEmpty()

	if len(data.ContactVendor) != 0 {
		for i := range data.ContactVendor {
			data.ContactVendor[i].Active = booldatatype.FromBool(true)
			data.ContactVendor[i].Position.SetNullIfEmpty()
			data.ContactVendor[i].Type.SetNullIfEmpty()
		}
	}

	if len(data.ItemCategoryVendor) != 0 {
		var temp []tblmastervendor.ItemCategoryVendor

		for i := range data.ItemCategoryVendor {
			if data.ItemCategoryVendor[i].ItemCategoryCode != "" {
				temp = append(temp, data.ItemCategoryVendor[i])
			}
		}

		data.ItemCategoryVendor = temp
		
		for i := range data.ItemCategoryVendor {
			data.ItemCategoryVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	if len(data.SectorVendor) != 0 {
		for i := range data.SectorVendor {
			data.SectorVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	if len(data.RatingVendor) != 0 {
		for i := range data.RatingVendor {
			data.RatingVendor[i].Active = booldatatype.FromBool(true)
		}
	}

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update vendor: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, nil
		}
		return nil, err
	}

	if len(res.ContactVendor) != 0 {
		var filteredContact []tblmastervendor.ContactVendor
		for _, detail := range res.ContactVendor {
			if detail.Active.ToBool(){
				filteredContact = append(filteredContact, detail)
			}
		}
		res.ContactVendor = filteredContact
	}

	if len(res.ItemCategoryVendor) != 0 {
		var filteredItCat []tblmastervendor.ItemCategoryVendor
		for _, detail := range res.ItemCategoryVendor {
			if detail.Active.ToBool(){
				filteredItCat = append(filteredItCat, detail)
			}
		}
		res.ItemCategoryVendor = filteredItCat
	}

	if len(res.SectorVendor) != 0 {
		var filteredSector []tblmastervendor.SectorVendor
		for _, detail := range res.SectorVendor {
			if detail.Active.ToBool(){
				filteredSector = append(filteredSector, detail)
			}
		}
		res.SectorVendor = filteredSector
	}

	if len(res.RatingVendor) != 0 {
		var filteredRating []tblmastervendor.RatingVendor
		for _, detail := range res.RatingVendor {
			if detail.Active.ToBool() {
				filteredRating = append(filteredRating, detail)
			}
		}
		res.RatingVendor = filteredRating
	}

	return res, nil
}

func (s *TblVendor) GetContact(ctx context.Context, code string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetContact(ctx, code, param)
}