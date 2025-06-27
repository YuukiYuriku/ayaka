package tblsite

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Read struct {
	Number     uint                      `json:"number"`
	SiteCode   string                    `db:"SiteCode" json:"site_code"`
	SiteName   string                    `db:"SiteName" json:"site_name"`
	Address    string                    `db:"Address" json:"address"`
	Active     booldatatype.BoolDataType `db:"Active" json:"active"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateDate string                    `db:"CreateDt" json:"create_date"`
}

type Create struct {
	SiteCode   string                    `db:"SiteCode" json:"site_code" validate:"required,whitespace,unique=tblsite->SiteCode,max=12"`
	SiteName   string                    `db:"SiteName" json:"site_name" validate:"required,max=50"`
	Address    string                    `db:"Address" json:"address" validate:"required,max=255"`
	Active     booldatatype.BoolDataType `db:"Active" json:"active" validate:"required"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark" validate:"max=255"`
	CreateDate string                    `db:"CreateDt" json:"create_date"`
	CreateBy   string                    `db:"CreateBy" json:"create_by"`
}

type Update struct {
	SiteCode       string                    `db:"SiteCode" json:"site_code" validate:"required,incolumn=tblsite->SiteCode,max=12"`
	SiteName       string                    `db:"SiteName" json:"site_name" validate:"required,max=50"`
	Address        string                    `db:"Address" json:"address" validate:"required,max=255"`
	Active         booldatatype.BoolDataType `db:"Active" json:"active" validate:"required"`
	Remark         nulldatatype.NullDataType `db:"Remark" json:"remark"`
	LastUpdateDate string                    `json:"create_date"`
	LastUpdateBy   string                    `json:"create_by"`
}
