package tblstocksummary

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type GetItem struct {
	Number   uint    `json:"number"`
	ItemCode string  `db:"ItCode" json:"item_code"`
	ItemName string  `db:"ItName" json:"item_name"`
	Batch    string  `db:"BatchNo" json:"batch"`
	Stock    float32 `db:"Stock" json:"stock"`
	UomName  string  `db:"UomName" json:"uom_name"`
}

type Fetch struct {
	Number        uint                      `json:"number"`
	WarehouseName string                    `db:"WhsName" json:"warehouse_name"`
	ItemCode      string                    `db:"ItCode" json:"item_code"`
	LocalCode     nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code"`
	ItemName      string                    `db:"ItName" json:"item_name"`
	Catgory       string                    `db:"ItCtName" json:"item_category_name"`
	Active        booldatatype.BoolDataType `db:"ActInd" json:"active"`
	Quantity      float32                   `db:"Stock" json:"quantity"`
	Uom           string                    `db:"UomName" json:"uom"`
}
