package tblmastervendor

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

// Read & Detail
type ContactVendorDetail struct {
	VendorCode string                    `db:"VendorCode" json:"vendor_code"`
	DNo        string                    `db:"DNo" json:"detail_no"`
	Name       string                    `db:"Name" json:"name"`
	Number     string                    `db:"Number" json:"number"`
	Position   nulldatatype.NullDataType `db:"Position" json:"position"`
	Type       nulldatatype.NullDataType `db:"Type" json:"type"`
}

type ItemCategoryVendorDetail struct {
	VendorCode       string `db:"VendorCode" json:"vendor_code"`
	ItemCategoryCode string `db:"ItCatCode" json:"item_category_code"`
}

type SectorVendorDetail struct {
	VendorCode       string `db:"VendorCode" json:"vendor_code"`
	VendorSectorCode string `db:"VendorSectorCode" json:"vendor_sector_code"`
	DNoVendorSector  string `db:"DNoVendorSector" json:"detail_vendor_sector"`
}

type RatingVendorDetail struct {
	VendorCode       string  `db:"VendorCode" json:"vendor_code"`
	VendorRatingCode string  `db:"VendorRatingCode" json:"vendor_rating_code"`
	Value            float32 `db:"Value" json:"value"`
}

type Read struct {
	Number     uint   `json:"number"`
	VendorCode string `db:"VendorCode" json:"vendor_code"`
	VendorName string `db:"VendorName" json:"vendor_name"`
	Address    string `db:"Address" json:"address"`
	CityName   string `db:"CityName" json:"city_name"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type Detail struct {
	VendorCode         string                     `db:"VendorCode" json:"vendor_code"`
	VendorName         string                     `db:"VendorName" json:"vendor_name"`
	VendorCatCode      string                     `db:"VendorCatCode" json:"vendor_category_code"`
	Address            nulldatatype.NullDataType  `db:"Address" json:"address"`
	CityCode           string                     `db:"CityCode" json:"city_code"`
	PostalCode         nulldatatype.NullDataType  `db:"PostalCode" json:"postal_code"`
	Website            nulldatatype.NullDataType  `db:"Website" json:"website"`
	HeadOffice         nulldatatype.NullDataType  `db:"HeadOffice" json:"head_office"`
	Phone              nulldatatype.NullDataType  `db:"Phone" json:"phone"`
	Mobile             nulldatatype.NullDataType  `db:"Mobile" json:"mobile"`
	Email              nulldatatype.NullDataType  `db:"Email" json:"email"`
	Remark             nulldatatype.NullDataType  `db:"Remark" json:"remark"`
	ContactVendor      []ContactVendorDetail      `json:"contact_vendor"`
	ItemCategoryVendor []ItemCategoryVendorDetail `json:"item_category_vendor"`
	SectorVendor       []SectorVendorDetail       `json:"sector_vendor"`
	RatingVendor       []RatingVendorDetail       `json:"rating_vendor"`
}

// Create & Update
type ContactVendor struct {
	VendorCode string                    `db:"VendorCode" json:"vendor_code"`
	DNo        string                    `db:"DNo" json:"detail_no"`
	Name       string                    `db:"Name" json:"name" validate:"max=255"`
	Number     string                    `db:"Number" json:"number" validate:"max=20"`
	Active     booldatatype.BoolDataType `db:"Active" json:"active"` 
	Position   nulldatatype.NullDataType `db:"Position" json:"position" validate:"max=50"`
	Type       nulldatatype.NullDataType `db:"Type" json:"type" validate:"max=50"`
}

type ItemCategoryVendor struct {
	VendorCode       string                    `db:"VendorCode" json:"vendor_code"`
	ItemCategoryCode string                    `db:"ItCatCode" json:"item_category_code" validate:"incolumn=tblitemcategory->ItCtCode"`
	Active           booldatatype.BoolDataType `db:"Active" json:"active"`
}

type SectorVendor struct {
	VendorCode       string                    `db:"VendorCode" json:"vendor_code"`
	VendorSectorCode string                    `db:"VendorSectorCode" json:"vendor_sector_code" validate:"incolumn=tblvendorsectordtl->SectorCode"`
	DNoVendorSector  string                    `db:"DNoVendorSector" json:"detail_vendor_sector"`
	Active           booldatatype.BoolDataType `db:"Active" json:"active"`
}

type RatingVendor struct {
	VendorCode       string                    `db:"VendorCode" json:"vendor_code"`
	VendorRatingCode string                    `db:"VendorRatingCode" json:"vendor_rating_code" validate:"incolumn=tblvendorrating->IndicatorCode"`
	Value            float32                   `db:"Value" validate:"min=0,max=5"`
	Active           booldatatype.BoolDataType `db:"Active" json:"active"`
}

type Create struct {
	VendorCode         string                     `json:"vendor_code"`
	VendorName         string                     `json:"vendor_name" validate:"required,max=255"`
	VendorCatCode      string                     `json:"vendor_category_code" validate:"required,incolumn=tblvendorcategory->VendorCatCode"`
	Address            nulldatatype.NullDataType  `json:"address" validate:"max=255"`
	CityCode           string                     `json:"city_code" validate:"required,incolumn=tblcity->CityCode"`
	PostalCode         nulldatatype.NullDataType  `json:"postal_code" validate:"max=10"`
	Website            nulldatatype.NullDataType  `json:"website" validate:"max=50"`
	HeadOffice         nulldatatype.NullDataType  `json:"head_office" validate:"max=60"`
	Phone              nulldatatype.NullDataType  `json:"phone" validate:"max=20"`
	Mobile             nulldatatype.NullDataType  `json:"mobile" validate:"max=20"`
	Email              nulldatatype.NullDataType  `json:"email" validate:"max=255"`
	Remark             nulldatatype.NullDataType  `json:"remark" validate:"max=255"`
	ContactVendor      []ContactVendor      `json:"contact_vendor" validate:"dive"`
	ItemCategoryVendor []ItemCategoryVendor `json:"item_category_vendor" validate:"dive"`
	SectorVendor       []SectorVendor       `json:"sector_vendor" validate:"dive"`
	RatingVendor       []RatingVendor       `json:"rating_vendor" validate:"dive"`
	CreateBy           string                     `json:"create_by"`
	CreateDate         string                     `json:"create_date"`
}

type Update struct {
	VendorCode         string                    `json:"vendor_code" validate:"required,incolumn=tblvendorhdr->VendorCode"`
	VendorName         string                    `json:"vendor_name" validate:"required,max=255"`
	VendorCatCode      string                    `json:"vendor_category_code" validate:"required,incolumn=tblvendorcategory->VendorCatCode"`
	Address            nulldatatype.NullDataType `json:"address" validate:"max=255"`
	CityCode           string                    `json:"city_code" validate:"required,incolumn=tblcity->CityCode"`
	PostalCode         nulldatatype.NullDataType `json:"postal_code" validate:"max=10"`
	Website            nulldatatype.NullDataType `json:"website" validate:"max=50"`
	HeadOffice         nulldatatype.NullDataType `json:"head_office" validate:"max=60"`
	Phone              nulldatatype.NullDataType `json:"phone" validate:"max=20"`
	Mobile             nulldatatype.NullDataType `json:"mobile" validate:"max=20"`
	Email              nulldatatype.NullDataType `json:"email" validate:"max=255"`
	Remark             nulldatatype.NullDataType `json:"remark" validate:"max=255"`
	ContactVendor      []ContactVendor           `json:"contact_vendor" validate:"dive"`
	ItemCategoryVendor []ItemCategoryVendor      `json:"item_category_vendor" validate:"dive"`
	SectorVendor       []SectorVendor            `json:"sector_vendor" validate:"dive"`
	RatingVendor       []RatingVendor            `json:"rating_vendor" validate:"dive"`
	LastUpdateBy       string                    `json:"last_update_by"`
	LastUpdateDate     string                    `json:"last_update_date"`
}
