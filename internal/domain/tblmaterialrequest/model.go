package tblmaterialrequest

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo          string                    `db:"DocNo" json:"document_number"`
	DNo            string                    `db:"DNo" json:"detail_number"`
	Cancel         booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	Success        booldatatype.BoolDataType `db:"SuccessInd" json:"success"`
	ItCode         string                    `db:"ItCode" json:"item_code"`
	ItName         string                    `db:"ItName" json:"item_name"`
	UomName        string                    `db:"UomName" json:"uom_name"`
	Qty            float32                   `db:"Qty" json:"quantity"`
	UsageDt        string                    `db:"UsageDt" json:"usage_date"`
	CurCode        string                    `db:"CurCode" json:"currency_code"`
	EstimatedPrice float32                   `db:"EstimatedPrice" json:"estimated_price"`
	Remark         nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Duration       float32                   `db:"Duration" json:"duration"`
	DurationUom    nulldatatype.NullDataType `db:"DurationUom" json:"duration_uom"`
}

type Create struct {
	DocNo      string                    `db:"DocNo" json:"document_number"`
	Date       string                    `db:"DocDt" json:"date" validate:"required"`
	SiteCode   nulldatatype.NullDataType `db:"SiteCode" json:"site_code"`
	Department string                    `db:"Department" json:"department" validate:"required"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details    []Detail                  `json:"details"`
	CreateBy   string                    `db:"CreateBy" json:"create_by"`
	CreateDt   string                    `db:"CreateDt" json:"create_date"`
}

type Read struct {
	Number     uint                      `json:"number"`
	DocNo      string                    `db:"DocNo" json:"document_number"`
	Date       string                    `db:"DocDt" json:"date"`
	TblDate    string                    `json:"table_date"`
	SiteCode   nulldatatype.NullDataType `db:"SiteCode" json:"site_code"`
	Department string                    `db:"Department" json:"department"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	Details    []Detail                  `json:"details"`
}

type GetMaterialRequest struct {
	Number         uint    `json:"number"`
	DocNo          string  `db:"MaterialReqDocNo" json:"material_request_document"`
	DNo            string  `db:"MaterialReqDNo" json:"material_request_detail"`
	Department     string  `db:"Department" json:"department"`
	Qty            float32 `db:"OutstandingQty" json:"quantity"`
	ItCode         string  `db:"ItCode" json:"item_code"`
	ItName         string  `db:"ItName" json:"item_name"`
	EstimatedPrice float32 `db:"EstimatedPrice" json:"estimated_price"`
	UsageDt        string  `db:"UsageDt" json:"usage_date"`
}

type OutstandingMaterial struct {
	Number         uint    `json:"number"`
	DocNo          string  `db:"DocNo" json:"document_number"`
	Department     string  `db:"Department" json:"department"`
	ItName         string  `db:"ItName" json:"item_name"`
	RequestedQty   float32 `db:"RequestedQty" json:"requested_quantity"`
	ReceivedQty    float32 `db:"ReceivedQty" json:"received_quantity"`
	OutstandingQty float32 `db:"OutstandingQty" json:"outstanding_quantity"`
	UomName        string  `db:"UomName" json:"uom_name"`
	UsageDate      string  `db:"UsageDt" json:"usage_date"`
}
