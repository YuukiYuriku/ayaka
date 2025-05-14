package tblcity

import (
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type ReadTblCity struct {
	Number     uint                      `json:"number"`
	CityCode   string                    `db:"CityCode" json:"city_code"`
	CityName   string                    `db:"CityName" json:"city_name"`
	ProvCode   string                    `db:"ProvCode" json:"province_code"`
	Province   string                    `db:"ProvName" json:"province"`
	RingArea   nulldatatype.NullDataType `db:"RingCode" json:"ring_area"`
	Location   nulldatatype.NullDataType `db:"LocationCode" json:"location"`
	CreateDate string                    `db:"CreateDt" json:"create_date"`
}

type CreateTblCity struct {
	CityCode   string                    `db:"CityCode" json:"city_code" validate:"required,min=1,max=16,unique=tblcity->CityCode" label:"City Code"`
	CityName   string                    `db:"CityName" json:"city_name" validate:"required,min=1,max=80" label:"City Name"`
	Province   string                    `db:"ProvCode" json:"province" validate:"required,incolumn=tblprovince->ProvCode" label:"Province"`
	RingArea   nulldatatype.NullDataType `db:"RingCode" json:"ring_area" validate:"max=5" label:"Ring Area"`
	Location   nulldatatype.NullDataType `db:"LocationCode" json:"location" validate:"max=40" label:"Location"`
	CreateBy   string                    `db:"CreateBy" json:"create_by"`
	CreateDate string                    `db:"CreateDt" json:"create_date"`
}

type UpdateTblCity struct {
	CityCode       string                    `db:"CityCode" json:"city_code" validate:"required,min=1,max=16,incolumn=tblcity->CityCode" label:"City Code"`
	CityName       string                    `db:"CityName" json:"city_name" validate:"required,min=1,max=80" label:"City Name"`
	Province       string                    `db:"ProvName" json:"province" validate:"required,incolumn=tblprovince->ProvCode" label:"Province"`
	RingArea       nulldatatype.NullDataType `db:"RingCode" json:"ring_area" validate:"max=5" label:"Ring Area"`
	Location       nulldatatype.NullDataType `db:"LocationCode" json:"location" validate:"max=40" label:"Location"`
	UserCode       string                    `db:"UserCode" json:"user_code"`
	LastUpdateDate string                    `db:"LastUpDt" json:"last_update_date"`
}

type GroupCityByProv struct {
	GroupedData string `db:"GroupedData"`
	ProvName    string `db:"ProvName"`
}
