package tblrecvvddtl

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DNo          string                    `db:"DNo" json:"dno"`
	Cancel       booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	CancelReason nulldatatype.NullDataType `db:"CancelReason" json:"cancel_reason"`
	ItemName     string                    `db:"ItName" json:"item_name"`
	LocalCode    nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code"`
	Batch        string                    `db:"BatchNo" json:"batch"`
	Price        float32                   `db:"UPrice" json:"price"`
	Quantity     float32                   `db:"QtyPurchase" json:"quantity"`
	UomName      string                    `db:"UomName" json:"uom_name"`
	Discount     float32                   `db:"Discount" json:"discount"`
	Rounding     float32                   `db:"RoundingValue" json:"rounding"`
}

type Create struct {
	DocNo        string                    `json:"document_number"`
	DNo          string                    `json:"dno"`
	Cancel       booldatatype.BoolDataType `json:"cancel"`
	CancelReason string                    `json:"cancel_reason"`
	ItemCode     string                    `json:"item_code"`
	LocalCode    string                    `json:"local_code"`
	Batch        string                    `json:"batch"`
	Price        float32                   `json:"price" validate:"required,min=1"`
	Quantity     float32                   `json:"quantity" validate:"required,min=1"`
	Discount     float32                   `json:"discount" validate:"min=0"`
	Rounding     float32                   `json:"rounding" validate:"min=0"`
	Remark       string                    `json:"remark"`
	Expired      string                    `json:"expired"`
}
