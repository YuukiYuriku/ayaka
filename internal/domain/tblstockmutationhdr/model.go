package tblstockmutationhdr

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	"gitlab.com/ayaka/internal/domain/tblstockmutationdtl"
)

type Fetch struct {
	Number    uint                      `json:"number"`
	DocNo     string                    `db:"DocNo" json:"document_number"`
	Cancel    booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	Date      string                    `db:"DocDt" json:"date"`
	TblDate   string 					`json:"table_date"`
	Warehouse string                    `db:"WhsName" json:"warehouse"`
	FromTo    string                    `db:"FromTo" json:"from_to"`
	ItemName  string                    `db:"ItName" json:"item_name"`
	Batch     string                    `db:"BatchNo" json:"batch"`
	Quantity  string                    `db:"Qty" json:"quantity"`
	Uom       string                    `db:"UomName" json:"uom"`
}

type Detail struct {
	DocNo         string                       `db:"DocNo" json:"document_number" validate:"incolumn=tblstockmutationhdr->DocNo"`
	Date          string                       `db:"DocDt" json:"date"`
	WarehouseCode string                       `db:"WhsCode" json:"warehouse_code"`
	WarehouseName string                       `db:"WhsName" json:"warehouse_name"`
	CancelReason  nulldatatype.NullDataType    `db:"CancelReason" json:"cancel_reason"`
	Cancel        booldatatype.BoolDataType    `db:"CancelInd" json:"cancel"`
	Remark        nulldatatype.NullDataType    `db:"Remark" json:"remark"`
	FromArray     []tblstockmutationdtl.Detail `db:"from_array" json:"from_array"`
	ToArray       []tblstockmutationdtl.Detail `db:"to_array" json:"to_array"`
}

type Create struct {
	DocNo         string                       `json:"document_number"`
	DocDate       string                       `json:"date" validate:"required"`
	WarehouseCode string                       `json:"warehouse_code" validate:"required,incolumn=tblwarehouse->WhsCode" label:"Warehouse"`
	BatchNo       string                       `json:"batch"`
	Source        string                       `json:"source"`
	Remark        nulldatatype.NullDataType    `json:"remark"`
	CreateBy      string                       `json:"create_by"`
	CreateDate    string                       `json:"create_date"`
	FromArray     []tblstockmutationdtl.Create `json:"from_array" validate:"dive"`
	ToArray       []tblstockmutationdtl.Create `json:"to_array" validate:"dive"`
}
