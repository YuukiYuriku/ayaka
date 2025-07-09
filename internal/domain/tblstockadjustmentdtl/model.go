package tblstockadjustmentdtl

import "gitlab.com/ayaka/internal/domain/shared/nulldatatype"

type Detail struct {
	DocNo         string                    `db:"DocNo" json:"document_number"`
	DNo           string                    `db:"DNo" json:"d_no"`
	ItemName      string                    `db:"ItName" json:"item_name"`
	Batch         string                    `db:"BatchNo" json:"batch"`
	StockSystem   float32                   `db:"Qty" json:"stock_system"`
	StockActual   float32                   `db:"QtyActual" json:"stock_actual"`
	Balance       float32                   `db:"Balance" json:"balance"`
	Uom           string                    `db:"UomName" json:"uom_name"`
	Spesification nulldatatype.NullDataType `db:"Specification" json:"specification"`
}

type Create struct {
	DocNo       string  `db:"DocNo" json:"document_number"`
	DNo         string  `db:"DNo" json:"d_no"`
	ItemCode    string  `db:"ItCode" json:"item_code" validate:"incolumn=tblitem->ItCode"`
	Batch       string  `db:"BatchNo" json:"batch"`
	StockSystem float32 `db:"Qty" json:"stock_system"`
	StockActual float32 `db:"QtyActual" json:"stock_actual" validate:"min=1" label:"Actual Stock"`
	Source      string  `db:"Source"`
}
