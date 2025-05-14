package tblwarehousecategory

import "gitlab.com/ayaka/internal/domain/logactivity"

type ReadTblWarehouseCategory struct {
	Number     uint   `json:"number"`
	WhsCtCode  string `db:"WhsCtCode" json:"warehouse_category_code"`
	WhsCtName  string `db:"WhsCtName" json:"warehouse_category_name"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type DetailTblWarehouseCategory struct {
	WhsCtCode   string                     `db:"WhsCtCode" json:"warehouse_category_code"`
	WhsCtName   string                     `db:"WhsCtName" json:"warehouse_category_name"`
	Remark      string                     `db:"Remark" json:"remark"`
	CreateBy    string                     `db:"CreateBy" json:"create_by"`
	CreateDate  string                     `db:"CreateDt" json:"create_date"`
	LogActivity []*logactivity.LogActivity `json:"log_activity"`
}

type CreateTblWarehouseCategory struct {
	WhsCtCode  string `db:"WhsCtCode" json:"warehouse_category_code" validate:"required,unique=tblwarehousecategory->WhsCtCode" label:"Warehouse Category Code"`
	WhsCtName  string `db:"WhsCtName" json:"warehouse_category_name" validate:"required,unique=tblwarehousecategory->WhsCtName" label:"Warehouse Category Name"`
	Remark     string `db:"Remark" json:"remark"`
	CreateBy   string `db:"CreateBy" json:"create_by"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type UpdateTblWarehouseCategory struct {
	WhsCtCode      string `db:"WhsCtCode" json:"warehouse_category_code" validate:"required,incolumn=tblwarehousecategory->WhsCtCode" label:"Warehouse Category Code"`
	WhsCtName      string `db:"WhsCtName" json:"warehouse_category_name" validate:"unique=tblwarehousecategory->WhsCtName" label:"Warehouse Category Name"`
	Remark         string `db:"Remark" json:"remark"`
	UserCode       string `db:"UserCode" json:"user_code"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date"`
	LastUpdateBy   string `db:"LastUpBy" json:"last_update_by"`
}
