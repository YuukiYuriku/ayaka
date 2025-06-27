package tblvendorsector

import "gitlab.com/ayaka/internal/domain/shared/booldatatype"

type VendorSectorDtl struct {
	SectorCode    string                    `db:"SectorCode" json:"sector_code"`
	DNo           string                    `db:"DNo" json:"detail_vendor_sector"`
	SubSectorName string                    `db:"SubSectorName" validate:"required,max=255" json:"sub_sector_name"`
	Active        booldatatype.BoolDataType `db:"Active"`
}

type Read struct {
	Number     uint                      `json:"number"`
	SectorCode string                    `db:"SectorCode" json:"sector_code"`
	SectorName string                    `db:"SectorName" json:"sector_name"`
	Active     booldatatype.BoolDataType `db:"Active" json:"active"`
	CreateDate string                    `db:"CreateDt" json:"create_date"`
	Details    []VendorSectorDtl         `json:"details"`
}

type Create struct {
	SectorCode string                    `json:"sector_code" validate:"required,whitespace,unique=tblvendorsectorhdr->SectorCode,max=60"`
	SectorName string                    `json:"sector_name" validate:"required,max=255"`
	Active     booldatatype.BoolDataType `json:"active" validate:"required"`
	CreateDate string                    `json:"create_date"`
	CreateBy   string                    `json:"create_by"`
	Details    []VendorSectorDtl         `json:"details" validate:"dive"`
}

type Update struct {
	SectorCode     string                    `json:"sector_code" validate:"required,whitespace,incolumn=tblvendorsectorhdr->SectorCode"`
	SectorName     string                    `json:"sector_name" validate:"required,max=255"`
	Active         booldatatype.BoolDataType `json:"active" validate:"required"`
	LastUpdateDate string                    `json:"last_update_date"`
	LastUpdateBy   string                    `json:"last_update_by"`
	Details        []VendorSectorDtl         `json:"details" validate:"dive"`
}

type GetSector struct {
	SectorCode string `db:"SectorCode" json:"sector_code"`
	SectorName string `db:"SectorName" json:"sector_name"`
}

type GetSubSector struct {
	DNo string `db:"DNo" json:"detail_vendor_sector"`
	SubSectorName string `db:"SubSectorName" json:"sub_sector_name"`
}