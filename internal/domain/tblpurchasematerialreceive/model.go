package tblpurchasematerialreceive

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo              string                    `db:"DocNo" json:"document_number"`
	DNo                string                    `db:"DNo" json:"detail_number"`
	CancelInd          booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	PurchaseOrderDocNo string                    `db:"PurchaseOrderDocNo" json:"purchase_order_document"`
	PurchaseOrderDNo   string                    `db:"PurchaseOrderDNo" json:"purchase_order_detail"`
	ItCode             string                    `db:"ItCode" json:"item_code"`
	ItName             string                    `db:"ItName" json:"item_name"`
	UomName            string                    `db:"UomName" json:"uom_name"`
	BatchNo            string                    `db:"BatchNo" json:"batch"`
	Source             string                    `db:"Source" json:"source"`
	OutstandingQty     float32                   `db:"OutstandingQty" json:"outstanding_quantity"`
	PurchaseQty        float32                   `db:"PurchaseQty" json:"purchase_quantity"`
	InventoryQty       float32                   `db:"InventoryQty" json:"inventory_quantity"`
	Remark             nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Read struct {
	Number     uint                      `json:"number"`
	DocNo      string                    `db:"DocNo" json:"document_number"`
	Date       string                    `db:"DocDt" json:"date"`
	TblDate    string                    `json:"table_date"`
	WhsCode    string                    `db:"WhsCode" json:"warehouse_code"`
	WhsName    string                    `db:"WhsName" json:"warehouse_name"`
	VendorCode string                    `db:"VendorCode" json:"vendor_code"`
	VendorName string                    `db:"VendorName" json:"vendor_name"`
	SiteCode   nulldatatype.NullDataType `db:"SiteCode" json:"site_code"`
	SiteName   nulldatatype.NullDataType `db:"SiteName" json:"site_name"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details    []Detail                  `json:"details"`
}

type Create struct {
	DocNo      string                    `db:"DocNo" json:"document_number"`
	Date       string                    `db:"DocDt" json:"date" validate:"required"`
	WhsCode    string                    `db:"WhsCode" json:"warehouse_code" validate:"required"`
	VendorCode string                    `db:"VendorCode" json:"vendor_code" validate:"required"`
	SiteCode   nulldatatype.NullDataType `db:"SiteCode" json:"site_code"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details    []Detail                  `json:"details"`
	CreateBy   string                    `db:"CreateBy" json:"created_by"`
	CreateDt   string                    `db:"CreateDt" json:"created_date"`
}

type Reporting struct {
	Number         uint                      `json:"number"`
	DocNo          string                    `db:"DocNo" json:"document_number"`
	Date           string                    `db:"DocDt" json:"date"`
	WhsName        string                    `db:"WhsName" json:"warehouse_name"`
	PODoc          string                    `db:"PODoc" json:"purchase_document"`
	VendorName     string                    `db:"VendorName" json:"vendor_name"`
	ItName         string                    `db:"ItName" json:"item_name"`
	BatchNo        string                    `db:"BatchNo" json:"batch"`
	Quantity       float32                   `db:"Qty" json:"quantity"`
	UomName        string                    `db:"UomName" json:"uom_name"`
	DocumentRemark nulldatatype.NullDataType `db:"DocumentRemark" json:"document_remark"`
	ItemRemark     nulldatatype.NullDataType `db:"ItemRemark" json:"item_remark"`
}
