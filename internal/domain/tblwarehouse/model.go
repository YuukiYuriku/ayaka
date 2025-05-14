package tblwarehouse

import (
	"gitlab.com/ayaka/internal/domain/logactivity"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type ReadTblWarehouse struct {
	Number        uint                      `json:"number"`
	WhsCode       string                    `db:"WhsCode" json:"warehouse_code" label:"Warehouse Code"`
	WhsName       string                    `db:"WhsName" json:"warehouse_name" label:"Warehouse Name"`
	WhsCtCode     string                    `db:"WhsCtCode" json:"warehouse_category_code" label:"Warehouse Category Code"`
	WhsCtName     string                    `db:"WhsCtName" json:"warehouse_category_name"`
	CityCode      string                    `db:"CityCode" json:"city_code" label:"City Code"`
	CityName      string                    `db:"CityName" json:"city_name"`
	PostalCd      nulldatatype.NullDataType `db:"PostalCd" json:"postal_cd" label:"Postal Code"`
	Phone         nulldatatype.NullDataType `db:"Phone" json:"phone" label:"Phone"`
	Fax           nulldatatype.NullDataType `db:"Fax" json:"fax" label:"Fax"`
	Email         nulldatatype.NullDataType `db:"Email" json:"email" label:"Email"`
	Mobile        nulldatatype.NullDataType `db:"Mobile" json:"mobile" label:"Mobile"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark" label:"Remark"`
	ContactPerson nulldatatype.NullDataType `db:"ContactPerson" json:"contact_person" label:"Contact Person"`
	CreateDate    string                    `db:"CreateDt" json:"create_date" label:"Creation Date"`
}

type DetailTblWarehouse struct {
	WhsCode       string                     `db:"WhsCode" json:"warehouse_code" label:"Warehouse Code"`
	WhsName       string                     `db:"WhsName" json:"warehouse_name" label:"Warehouse Name"`
	WhsCtCode     string                     `db:"WhsCtCode" json:"warehouse_category_code" label:"Warehouse Category Code"`
	WhsCtName     string                     `db:"WhsCtName" json:"warehouse_category_name"`
	Address       string                     `db:"Address" json:"address" label:"Address"`
	CityCode      string                     `db:"CityCode" json:"city_code" label:"City Code"`
	CityName      string                     `db:"CityName" json:"city_name"`
	PostalCd      string                     `db:"PostalCd" json:"postal_cd" label:"Postal Code"`
	Phone         string                     `db:"Phone" json:"phone" label:"Phone"`
	Fax           string                     `db:"Fax" json:"fax" label:"Fax"`
	Email         string                     `db:"Email" json:"email" label:"Email"`
	Mobile        string                     `db:"Mobile" json:"mobile" label:"Mobile"`
	ContactPerson string                     `db:"ContactPerson" json:"contact_person" label:"Contact Person"`
	Remark        string                     `db:"Remark" json:"remark" label:"Remark"`
	CreateBy      string                     `db:"CreateBy" json:"create_by" label:"Created By"`
	CreateDate    string                     `db:"CreateDt" json:"create_date" label:"Creation Date"`
	LastUpBy      string                     `db:"LastUpBy" json:"last_up_by" label:"Last Updated By"`
	LastUpDt      string                     `db:"LastUpDt" json:"last_up_dt" label:"Last Update Date"`
	LogActivity   []*logactivity.LogActivity `json:"log_activity" label:"Log Activities"`
}

type CreateTblWarehouse struct {
	WhsCode       string `db:"WhsCode" json:"warehouse_code" validate:"required,unique=tblwarehouse->WhsCode" label:"Warehouse Code"`
	WhsName       string `db:"WhsName" json:"warehouse_name" validate:"required,unique=tblwarehouse->WhsName" label:"Warehouse Name"`
	WhsCtCode     string `db:"WhsCtCode" json:"warehouse_category_code" label:"Warehouse Category Code"`
	Address       string `db:"Address" json:"address" label:"Address"`
	CityCode      string `db:"CityCode" json:"city_code" label:"City Code"`
	PostalCd      string `db:"PostalCd" json:"postal_code" label:"Postal Code"`
	Phone         string `db:"Phone" json:"phone" label:"Phone"`
	Fax           string `db:"Fax" json:"fax" label:"Fax"`
	Email         string `db:"Email" json:"email" label:"Email"`
	Mobile        string `db:"Mobile" json:"mobile" label:"Mobile"`
	ContactPerson string `db:"ContactPerson" json:"contact_person" label:"Contact Person"`
	Remark        string `db:"Remark" json:"remark" label:"Remark"`
	CreateBy      string `db:"CreateBy" json:"create_by" label:"Created By"`
	CreateDate    string `db:"CreateDt" json:"create_date" label:"Creation Date"`
}

type UpdateTblWarehouse struct {
	WhsCode        string `db:"WhsCode" json:"warehouse_code" validate:"required,incolumn=tblwarehouse->WhsCode" label:"Warehouse Code"`
	WhsName        string `db:"WhsName" json:"warehouse_name" validate:"required" label:"Warehouse Name"`
	WhsCtCode      string `db:"WhsCtCode" json:"warehouse_category_code" label:"Warehouse Category Code"`
	Address        string `db:"Address" json:"address" label:"Address"`
	CityCode       string `db:"CityCode" json:"city_code" label:"City Code"`
	PostalCd       string `db:"PostalCd" json:"postal_cd" label:"Postal Code"`
	Phone          string `db:"Phone" json:"phone" label:"Phone"`
	Fax            string `db:"Fax" json:"fax" label:"Fax"`
	Email          string `db:"Email" json:"email" label:"Email"`
	Mobile         string `db:"Mobile" json:"mobile" label:"Mobile"`
	ContactPerson  string `db:"ContactPerson" json:"contact_person" label:"Contact Person"`
	Remark         string `db:"Remark" json:"remark" label:"Remark"`
	LastUpBy       string `db:"LastUpBy" json:"last_up_by" label:"Last Updated By"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date" label:"Last Update Date"`
}
