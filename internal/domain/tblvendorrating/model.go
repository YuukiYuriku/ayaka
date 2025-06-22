package tblvendorrating

import "gitlab.com/ayaka/internal/domain/shared/booldatatype"

type Read struct {
	Number        uint                      `json:"number"`
	IndicatorCode string                    `db:"IndicatorCode" json:"indicator_code"`
	Description   string                    `db:"Description" json:"description"`
	Active        booldatatype.BoolDataType `db:"ActiveInd" json:"active"`
	CreateDate    string                    `db:"CreateDt" json:"create_date"`
}

type Create struct {
	IndicatorCode string                    `validate:"required,whitespace,unique=tblvendorrating->IndicatorCode,max=50" json:"indicator_code"`
	Description   string                    `validate:"required,max=255" json:"description"`
	Active        booldatatype.BoolDataType `validate:"required" json:"active"`
	CreateDate    string                    `json:"create_date"`
	CreateBy      string                    `json:"create_by"`
}

type Update struct {
	IndicatorCode  string                    `validate:"required,whitespace,incolumn=tblvendorrating->IndicatorCode,max=50" json:"indicator_code"`
	Description    string                    `validate:"max=255" json:"description"`
	Active         booldatatype.BoolDataType `json:"active"`
	LastUpdateDate string                    `json:"last_update_date"`
	LastUpdateBy   string                    `json:"last_update_by"`
}
