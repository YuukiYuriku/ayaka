package tblpurchasereturndelivery

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo                        string                    `db:"DocNo" json:"document_number"`
	DNo                          string                    `db:"DNo" json:"detail_number"`
	CancelInd                    booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	PurchaseMaterialReceiveDocNO string                    `db:"PurchaseMaterialReceiveDocNo" json:"purchase_document" validate:"required"`
	PurchaseMaterialReceiveDNo   string                    `db:"PurchaseMaterialReceiveDNo" json:"purchase_detail" validate:"required"`
	Date                         string                    `db:"DocDt" json:"date"`
	ItCode                       string                    `db:"ItCode" json:"item_code" validate:"required"`
	ItName                       string                    `db:"ItName" json:"item_name"`
	BatchNo                      string                    `db:"BatchNo" json:"batch"`
	Source                       string                    `db:"Source" json:"source"`
	Stock                        float32                   `db:"Stock" json:"stock"`
	QtyPurchase                  float32                   `db:"QtyPurchase" json:"quantity_purchase"`
	Qty                          float32                   `db:"Qty" json:"quantity"`
	Remark                       nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Read struct {
	Number     uint                      `json:"number"`
	DocNo      string                    `db:"DocNo" json:"document_number" validate:"incolumn=tblpurchasereturndeliveryhdr->DocNo"`
	Date       string                    `db:"DocDt" json:"date"`
	TblDate    string                    `json:"table_date"`
	WhsCode    string                    `db:"WhsCode" json:"warehouse_code"`
	WhsName    string                    `db:"WhsName" json:"warehouse_name"`
	VendorCode string                    `db:"VendorCode" json:"vendor_code"`
	VendorName string                    `db:"VendorName" json:"vendor_name"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details    []Detail                  `db:"Details" json:"details"`
}

type Create struct {
	DocNo      string                    `db:"DocNo" json:"document_number"`
	Date       string                    `db:"DocDt" json:"date" validate:"required"`
	WhsCode    string                    `db:"WhsCode" json:"warehouse_code" validate:"required"`
	VendorCode string                    `db:"VendorCode" json:"vendor_code" validate:"required"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy   string                    `db:"CreateBy" json:"created_by"`
	CreateDt   string                    `db:"CreateDt" json:"created_date"`
	Details    []Detail                  `db:"Details" json:"details"`
}

type GetReturnMaterial struct {
	Number        uint    `json:"number"`
	PurchaseDocNo string  `db:"PurchaseDocNo" json:"purchase_document"`
	PurchaseDNo   string  `db:"PurchaseDNo" json:"purchase_detail"`
	Date          string  `db:"DocDt" json:"date"`
	ItCode        string  `db:"ItCode" json:"item_code"`
	ItName        string  `db:"ItName" json:"item_name"`
	BatchNo       string  `db:"BatchNo" json:"batch"`
	Source        string  `db:"Source" json:"source"`
	Stock         float32  `db:"Stock" json:"stock"`
	QtyPurchase   float32 `db:"QtyPurchase" json:"qty_purchase"`
}
