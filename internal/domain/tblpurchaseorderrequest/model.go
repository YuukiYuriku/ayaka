package tblpurchaseorderrequest

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo            string                    `db:"DocNo" json:"document_number"`
	DNo              string                    `db:"DNo" json:"detail_number"`
	CancelInd        booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	SuccessInd       booldatatype.BoolDataType `db:"SuccessInd" json:"success"`
	MaterialReqDocNo string                    `db:"MaterialReqDocNo" json:"material_request_document" validate:"incolumn=tblmaterialrequestdtl->DocNo"`
	MaterialReqDNo   string                    `db:"MaterialReqDNo" json:"material_request_detail" validate:"incolumn=tblmaterialrequestdtl->DocNo"`
	ItCode           string                    `db:"ItCode" json:"item_code"`
	ItName           string                    `db:"ItName" json:"item_name"`
	Department       string                    `db:"Department" json:"department"`
	Qty              float32                   `db:"Qty" json:"quantity"`
	Total            float32                   `db:"Total" json:"total"`
	UsageDt          string                    `db:"UsageDt" json:"usage_date"`
	VendorQTDocNo    string                    `db:"VendorQTDocNo" json:"vendor_quotation_document"`
	VendorQTDNo      string                    `db:"VendorQTDNo" json:"vendor_quotation_detail"`
	VendorCode       string                    `db:"VendorCode" json:"vendor_code"`
	VendorName       string                    `db:"VendorName" json:"vendor_name"`
	ActualCurName    string                    `db:"ActualCurName" json:"actual_currency_name"`
	ActualPrice      float32                   `db:"Price" json:"actual_price"`
	TermOfPayment    string                    `db:"TermOfPayment" json:"term_of_payment"`
	DeliveryType     nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
	Remark           nulldatatype.NullDataType `db:"Remark" json:"remark"`
	EstimatedPrice   float32                   `db:"EstimatedPrice" json:"estimated_price"`
}

type Create struct {
	DocNo    string                    `db:"DocNo" json:"document_number"`
	Date     string                    `db:"DocDt" json:"date" validate:"required"`
	TblDate  string                    `json:"table_date"`
	Remark   nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy string                    `json:"create_by"`
	CreateDt string                    `json:"create_date"`
	Details  []Detail                  `db:"Detail" json:"details"`
}

type Read struct {
	Number  uint                      `json:"number"`
	DocNo   string                    `db:"DocNo" json:"document_number"`
	Date    string                    `db:"DocDt" json:"date"`
	TblDate string                    `json:"table_date"`
	Remark  nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details []Detail                  `db:"Detail" json:"details"`
}

type GetPurchaseOrderRequest struct {
	Number       uint                      `json:"number"`
	DocNo        string                    `db:"DocNo" json:"purchase_order_request_document"`
	DNo          string                    `db:"DNo" json:"purchase_order_request_detail"`
	ItCode       string                    `db:"ItCode" json:"item_code"`
	ItName       string                    `db:"ItName" json:"item_name"`
	Qty          float32                   `db:"Qty" json:"quantity"`
	UomName      string                    `db:"UomName" json:"uom_name"`
	CurName      string                    `db:"CurName" json:"currency_name"`
	Price        float32                   `db:"Price" json:"unit_price"`
	DeliveryType nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
}
