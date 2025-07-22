package tblvendorquotation

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo     string                    `db:"DocNo" json:"document_number"`
	DNo       string                    `db:"DNo" json:"detail_number"`
	ItCode    string                    `db:"ItCode" json:"item_code" validate:"required"`
	ItName    string                    `db:"ItName" json:"item_name"`
	ActiveInd booldatatype.BoolDataType `db:"ActiveInd" json:"active"`
	UsedInd   booldatatype.BoolDataType `db:"UsedInd" json:"used"`
	Price     float64                   `db:"Price" json:"price" validate:"min=0"`
	Remark    nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy  string                    `db:"CreateBy" json:"create_by"`
	CreateDt  string                    `db:"CreateDt" json:"create_date"`
}

type Create struct {
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"date" validate:"required"`
	Status        string                    `db:"Status" json:"status"`
	VendorCode    string                    `db:"VendorCode" json:"vendor_code" validate:"required"`
	TermOfPayment string                    `db:"TermOfPayment" json:"term_of_payment" validate:"required"`
	CurCode       string                    `db:"CurCode" json:"currency_code" validate:"required"`
	DeliveryType  nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy      string                    `db:"CreateBy" json:"create_by"`
	CreateDt      string                    `db:"CreateDt" json:"create_date"`
	Details       []Detail                  `json:"details" validate:"required"`
}

type Read struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"date"`
	TblDate       string                    `json:"table_date"`
	Status        string                    `db:"Status" json:"status"`
	VendorCode    string                    `db:"VendorCode" json:"vendor_code"`
	VendorName    string                    `db:"VendorName" json:"vendor_name"`
	TermOfPayment string                    `db:"TermOfPayment" json:"term_of_payment"`
	CurCode       string                    `db:"CurCode" json:"currency_code"`
	DeliveryType  nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy      string                    `db:"CreateBy" json:"create_by"`
	CreateDt      string                    `db:"CreateDt" json:"create_date"`
	Details       []Detail                  `json:"details" validate:"required"`
}

type GetVendorQuotation struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"vendor_quotation_document"`
	DNo           string                    `db:"DNo" json:"vendor_quotation_detail"`
	VendorCode    string                    `db:"VendorCode" json:"vendor_code"`
	VendorName    string                    `db:"VendorName" json:"vendor_name"`
	TermOfPayment string                    `db:"TermOfPayment" json:"term_of_payment"`
	DeliveryType  nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
	CurName       string                    `db:"CurName" json:"actual_currency_name"`
	Price         float32                   `db:"Price" json:"actual_price"`
}
