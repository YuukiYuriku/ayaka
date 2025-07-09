package tbldirectpurchasercv

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo   string                    `db:"DocNo" json:"document_number"`
	DNo     string                    `db:"DNo" json:"detail_no"`
	Cancel  booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	ItCode  string                    `db:"ItCode" json:"item_code"`
	ItName  string                    `db:"ItName" json:"item_name"`
	UomName string                    `db:"UomName" json:"uom_name"`
	BatchNo string                    `db:"BatchNo" json:"batch"`
	Source  string                    `db:"Source" json:"source"`
	Price   float32                   `db:"Price" json:"price" validate:"min=0"`
	Qty     float32                   `db:"Qty" json:"quantity" validate:"min=0"`
}

type Create struct {
	DocNo         string                    `json:"document_number"`
	Date          string                    `json:"document_date" validate:"required"`
	Department    string                    `json:"department" validate:"required"`
	WhsCode       string                    `json:"warehouse_code" validate:"required"`
	VendorCode    string                    `json:"vendor_code" validate:"required"`
	SiteCode      nulldatatype.NullDataType `json:"site_code"`
	TermOfPayment string                    `json:"term_of_payment" validate:"required"`
	CurCode       string                    `json:"currency_code" validate:"required"`
	TaxCode       nulldatatype.NullDataType                    `json:"tax_code" validate:"incolumn=tbltax->TaxCode"`
	GrandTotal    float32                   `json:"grand_total"`
	Remark        nulldatatype.NullDataType `json:"remark"`
	CreateBy      string
	CreateDt      string
	Details       []Detail `json:"details" validate:"dive"`
}

type Read struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number" validate:"incolumn=tbldirectpurchasercvhdr->DocNo"`
	Date          string                    `db:"DocDt" json:"document_date"`
	TblDate       string                    `json:"table_date"`
	Department    string                    `db:"Department" json:"department"`
	WhsCode       string                    `db:"WhsCode" json:"warehouse_code"`
	WhsName       string                    `db:"WhsName" json:"warehouse_name"`
	VendorCode    string                    `db:"VendorCode" json:"vendor_code"`
	VendorName    string                    `db:"VendorName" json:"vendor_name"`
	SiteCode      nulldatatype.NullDataType `db:"SiteCode" json:"site_code"`
	TermOfPayment string                    `db:"TermOfPayment" json:"term_of_payment"`
	CurCode       string                    `db:"CurCode" json:"currency_code"`
	TaxCode       nulldatatype.NullDataType                    `db:"TaxCode" json:"tax_code"`
	TotalTax      float32                   `json:"total_tax"`
	TotalQuantity float32 `json:"total_quantity"`
	GrandTotal    float32                   `json:"grand_total"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details       []Detail                  `json:"details"`
}
