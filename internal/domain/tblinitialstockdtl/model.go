package tblinitialstockdtl

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Read struct {
	DocNo     string                    `db:"DocNo" json:"document_number"`
	DNo       string                    `db:"DNo" json:"d_no" validate:"incolumn=tblstockinitialdtl->DNo" label:"Doc Number"`
	Cancel    booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	ItemCode  string                    `db:"ItCode" json:"item_code,omitempty"`
	ItemName  string                    `db:"ItName" json:"item_name,omitempty"`
	LocalCode nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code,omitempty"`
	Batch     string                    `db:"BatchNo" json:"batch,omitempty"`
	Quantity  float32                   `db:"Qty" json:"quantity,omitempty"`
	Uom       string                    `db:"UomName" json:"uom,omitempty"`
	Price     float32                   `db:"UPrice" json:"price,omitempty"`
	Source    string                    `db:"Source" json:"source"`
}

type Create struct {
	DNo      string                    `db:"DNo" json:"d_no"`
	Cancel   booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	ItemCode string                    `db:"ItCode" json:"item_code" validate:"min=1,incolumn=tblitem->ItCode"`
	Batch    string                    `db:"BatchNo" json:"batch" validate:"max=250"`
	Quantity float32                   `db:"Qty" json:"quantity" validate:"min=1"`
	Price    float32                   `db:"UPrice" json:"price"`
	Source   string                    `db:"Source" json:"source"`
}
