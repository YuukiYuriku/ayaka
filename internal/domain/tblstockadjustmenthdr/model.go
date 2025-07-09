package tblstockadjustmenthdr

import (
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	"gitlab.com/ayaka/internal/domain/tblstockadjustmentdtl"
)

type Read struct {
	Number        uint   `json:"number"`
	DocNo         string `db:"DocNo" json:"document_number"`
	Date          string `db:"DocDt" json:"date"`
	WarehouseCode string `db:"WhsCode" json:"warehouse_code"`
	WarehouseName string `db:"WhsName" json:"warehouse_name"`
	TblDate		  string `json:"table_date"`
}

type Detail struct {
	Number        uint                           `json:"number,omitempty"`
	DocNo         string                         `db:"DocNo" json:"document_number,omitempty"`
	Date          string                         `db:"DocDt" json:"date,omitempty"`
	WarehouseCode string                         `db:"WhsCode" json:"warehouse_code,omitempty"`
	WarehouseName string                         `db:"WhsName" json:"warehouse_name,omitempty"`
	Remark        nulldatatype.NullDataType      `db:"Remark" json:"remark"`
	Details       []tblstockadjustmentdtl.Detail `db:"Detail" json:"details"`
	TotalBalance  float32                        `json:"total_balance"`
	TblDate       string 						 `json:"table_date"`
}

type Create struct {
	DocNo         string                         `db:"DocNo" json:"document_number"`
	Date          string                         `db:"DocDt" json:"date" validate:"required"`
	WarehouseCode string                         `db:"WhsCode" json:"warehouse_code" validate:"required,incolumn=tblwarehouse->WhsCode" label:"Warehouse"`
	Remark        nulldatatype.NullDataType      `db:"Remark" json:"remark"`
	Details       []tblstockadjustmentdtl.Create `json:"details" validate:"dive"`
	CreateBy      string                         `db:"CreateBy" json:"create_by"`
	CreateDate    string                         `db:"CreateDt" json:"create_date"`
}
