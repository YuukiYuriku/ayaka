package tblinitialstock

import (
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	"gitlab.com/ayaka/internal/domain/tblinitialstockdtl"
)

type Read struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"document_date"`
	WarehouseName string                    `db:"WhsName" json:"warehouse_name"`
	WarehouseCode string                    `db:"WhsCode" json:"warehouse_code,omitempty"`
	CurrencyCode  string                    `db:"CurCode" json:"currency_code,omitempty"`
	CurrencyName  string                    `db:"CurName" json:"currency_name,omitempty"`
	Rate          float32                   `db:"ExcRate" json:"rate,omitempty"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark,omitempty"`
	Details       []tblinitialstockdtl.Read `json:"details"`
	TotalQuantity float32                   `json:"total_quantity"`
}

type Detail struct {
	DocNo         string                    `db:"DocNo" json:"document_number,omitempty" validate:"incolumn=tblstockinitialhdr->DocNo" label:"Document Number"`
	WarehouseCode string                    `db:"WhsCode" json:"warehouse_code,omitempty"`
	CurrencyCode  string                    `db:"CurCode" json:"currency_code,omitempty"`
	CurrencyName  string                    `db:"CurName" json:"currency_name,omitempty"`
	Rate          float32                   `db:"ExcRate" json:"rate,omitempty"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark,omitempty"`
	Detail        []tblinitialstockdtl.Read `json:"details" validate:"dive"`
	TotalQuantity float32                   `json:"total_quantity"`
}

type Create struct {
	DocNo         string                      `db:"DocNo" json:"document_number"`
	Date          string                      `db:"DocDt" json:"document_date" validate:"required" label:"Document Date"`
	WarehouseCode string                      `db:"WhsCode" json:"warehouse_code" validate:"required,incolumn=tblwarehouse->WhsCode" label:"Warehouse"`
	CurrencyCode  string                      `db:"CurCode" json:"currency_code" validate:"required,incolumn=tblcurrency->CurCode" label:"Currency"`
	Rate          float32                     `db:"ExcRate" json:"rate" validate:"required"`
	Remark        nulldatatype.NullDataType   `db:"Remark" json:"remark" validate:"max=400"`
	CreateDate    string                      `db:"CreateDt" json:"create_date"`
	CreateBy      string                      `db:"CreateBy" json:"create_by"`
	DocType       string                      `db:"DocType" json:"doc_type"`
	Detail        []tblinitialstockdtl.Create `json:"details" validate:"dive"`
}
