package tbldailystockmovement

import "gitlab.com/ayaka/internal/domain/shared/nulldatatype"

type Read struct {
	Number        uint                      `json:"number"`
	ItemCode      string                    `db:"ItCode" json:"item_code"`
	ItemName      string                    `db:"ItName" json:"item_name"`
	Specification nulldatatype.NullDataType `db:"Specification" json:"specification"`
	Category      string                    `db:"ItCtName" json:"item_category_name"`
	Uom           string                    `db:"UomName" json:"uom_name"`
	Init          float32                   `db:"Qty" json:"initial_qty"`
	In            float32                   `db:"Qty2" json:"in_qty"`
	Out           float32                   `db:"Qty3" json:"out_qty"`
	Total         float32                   `db:"Total" json:"total_qty"`
	RealStock     float32                   `db:"ReakStock" json:"real_stock"`
}
