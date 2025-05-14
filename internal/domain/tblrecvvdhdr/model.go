package tblrecvvdhdr

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	"gitlab.com/ayaka/internal/domain/tblrecvvddtl"
)

type Fetch struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"document_date"`
	Cancel        booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	LocalCode     nulldatatype.NullDataType `db:"LocalDocNo" json:"local_code"`
	WarehouseName string                    `db:"WhsName" json:"warehouse_name"`
	Vendor        string                    `db:"VdCode" json:"vendor"`
	DO            nulldatatype.NullDataType `db:"VdDONo" json:"do"`
	ItemName      string                    `db:"ItName" json:"item_name"`
	LocalCodeItem string                    `db:"ItCodeInternal" json:"local_code_item"`
	ForeignName   string                    `db:"ForeignName" json:"foreign_name"`
	Batch         string                    `db:"BatchNo" json:"batch"`
	Lot           string                    `db:"Lot" json:"lot"`
	Bin           string                    `db:"Bin" json:"bin"`
	CurrencyCode  string                    `db:"CurCode" json:"currency_code"`
	CurrencyName  string                    `db:"CurName" json:"currency_name"`
	Price         float32                   `db:"UPrice" json:"price"`
	Quantity      float32                   `db:"QtyPurchase" json:"quantity"`
	UomCode       string                    `db:"PurchaseUOMCode" json:"uom_code"`
	UomName       string                    `db:"UomName" json:"uom_name"`
	Discount      float32                   `db:"Discount" json:"discount"`
	Rounding      float32                   `db:"RoundingValue" json:"rounding"`
	Amount        float32                   `db:"Amount" json:"amount"`
}

type Detail struct {
	DocNo          string                    `db:"DocNo" json:"document_number"`
	Date           string                    `db:"DocDt" json:"document_date"`
	LocalCode      nulldatatype.NullDataType `db:"LocalDocNo" json:"local_code"`
	WarehouseCode  string                    `db:"WhsCode" json:"warehouse_code"`
	WarehouseName  string                    `db:"WhsName" json:"warehouse_name"`
	Vendor         string                    `db:"VdCode" json:"vendor"`
	DO             nulldatatype.NullDataType `db:"VdDONo" json:"do"`
	Site           string                    `db:"SiteCode" json:"site"`
	TermOfPayment  string                    `db:"PtCode" json:"term_of_payment"`
	CurrerncyCode  string                    `db:"CurCode" json:"currency_code"`
	CurrencyName   string                    `db:"CurName" json:"currency_name"`
	Tax            float32                   `db:"TaxCode1" json:"tax"`
	TotalTax       float32                   `db:"TotalTax" json:"total_tax"`
	DiscountAmount float32                   `db:"DiscountAmt" json:"discount_amount"`
	GrandTotal     float32                   `json:"grand_total"`
	Remark         nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Project        nulldatatype.NullDataType `db:"ProjectCode" json:"project"`
	TotalQuantity  float32                   `json:"total_quantity"`
	Detail         []tblrecvvddtl.Detail     `json:"details"`
}

type Create struct {
	DocNo         string                    `json:"document_number"`
	Date          string                    `json:"document_date" validate:"required"`
	LocalCode     nulldatatype.NullDataType `json:"local_code"`
	Departement   string                    `json:"departement"`
	WarehouseCode string                    `json:"warehouse_code" validate:"required"`
	Vendor        string                    `json:"vendor"`
	DO            string                    `json:"do" validate:"max=30"`
	Site          nulldatatype.NullDataType `json:"site" validate:"max=16"`
	TermOfPayment string                    `json:"term_of_payment" validate:"max=16"`
	CurrencyCode  string                    `json:"currency_code" validate:"required"`
	Tax           float32                   `json:"tax"`
	Remark        nulldatatype.NullDataType `json:"remark"`
	Project       nulldatatype.NullDataType `json:"project" validate:"max=16"`
	Detail        []tblrecvvddtl.Create     `json:"details" validate:"required,dive"`
	CreateBy      string                    `json:"create_by"`
	CreateDate    string                    `json:"create_date"`
}
