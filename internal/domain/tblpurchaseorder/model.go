package tblpurchaseorder

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo                 string                    `db:"DocNo" json:"document_number"`
	DNo                   string                    `db:"DNo" json:"detail_number"`
	CancelInd             booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	SuccessInd            booldatatype.BoolDataType `db:"SuccessInd" json:"success"`
	PurchaseOrderReqDocNo string                    `db:"PurchaseOrderReqDocNo" json:"purchase_order_request_document" validate:"incolumn=tblpurchaseorderreqhdr->DocNo"`
	PurchaseOrderReqDNo   string                    `db:"PurchaseOrderReqDNo" json:"purchase_order_request_detail" validate:"incolumn=tblpurchaseorderreqdtl->DNo"`
	ItCode                string                    `db:"ItCode" json:"item_code" validate:"required"`
	ItName                string                    `db:"ItName" json:"item_name"`
	Qty                   float32                   `db:"Qty" json:"quantity" validate:"required"`
	CurName               string                    `db:"CurName" json:"currency_name"`
	UPrice                float32                   `db:"Price" json:"unit_price"`
	Total                 float32                   `db:"Total" json:"total"`
	DeliveryType          nulldatatype.NullDataType `db:"DeliveryType" json:"delivery_type"`
	Remark                nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Create struct {
	DocNo            string                    `db:"DocNo" json:"document_number"`
	Date             string                    `db:"DocDt" json:"date" validate:"required"`
	TblDate          string                    `json:"table_date"`
	Status           string                    `db:"Status" json:"status"`
	VendorCode       string                    `db:"VendorCode" json:"vendor_code" validate:"required"`
	ContactPersonDNo string                    `db:"DNo" json:"detail_number_contact_vendor"`
	Remark           nulldatatype.NullDataType `db:"Remark" json:"remark"`
	TaxCode          nulldatatype.NullDataType `db:"TaxCode" json:"tax_code"`
	TaxRate          float32                   `db:"TaxRate" json:"tax_rate"`
	CreateBy         string                    `json:"create_by"`
	CreateDt         string                    `db:"create_date"`
	Details          []Detail                  `json:"details" validate:"dive"`
}

type Read struct {
	Number           uint                      `json:"number"`
	DocNo            string                    `db:"DocNo" json:"document_number"`
	Date             string                    `db:"DocDt" json:"date"`
	TblDate          string                    `json:"table_date"`
	Status           string                    `db:"Status" json:"status"`
	VendorCode       string                    `db:"VendorCode" json:"vendor_code"`
	VendorName       string                    `db:"VendorName" json:"vendor_name"`
	ContactPersonDNo string                    `db:"DNo" json:"detail_number_contact_vendor"`
	Remark           nulldatatype.NullDataType `db:"Remark" json:"remark"`
	TaxCode          nulldatatype.NullDataType `db:"TaxCode" json:"tax_code"`
	TaxRate          float32                   `db:"TaxRate" json:"tax_rate"`
	TotalTax         float32                   `json:"total_tax"`
	GrandTotal       float32                   `json:"grand_total"`
	Details          []Detail                  `json:"details"`
}

type GetPurchaseOrder struct {
	Number         uint    `json:"number"`
	DocNo          string  `db:"DocNo" json:"purchase_order_document"`
	DNo            string  `db:"DNo" json:"purchase_order_detail"`
	Date           string  `db:"DocDt" json:"date"`
	ItCode         string  `db:"ItCode" json:"item_code"`
	ItName         string  `db:"ItName" json:"item_name"`
	OutstandingQty float32 `db:"OutstandingQty" json:"outstanding_quantity"`
	UomName        string  `db:"UomName" json:"uom_name"`
}

type OutstandingPO struct {
	Number         uint    `json:"number"`
	DocNo          string  `db:"DocNo" json:"document_number"`
	Status         string  `db:"Status" json:"status"`
	VendorName     string  `db:"VendorName" json:"vendor_name"`
	Department     string  `db:"Department" json:"department"`
	ItName         string  `db:"ItName" json:"item_name"`
	PurchaseQty    float32 `db:"PurchaseQty" json:"purchase_quantity"`
	OutstandingQty float32 `db:"OutstandingQty" json:"outstanding_quantity"`
	UomName        string  `db:"UomName" json:"uom_name"`
	CurName        string  `db:"CurName" json:"currency_name"`
	Price          float32 `db:"Price" json:"price"`
	Total          float32 `db:"Total" json:"total"`
}
