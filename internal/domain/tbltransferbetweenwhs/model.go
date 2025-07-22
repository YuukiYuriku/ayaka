package tbltransferbetweenwhs

import "gitlab.com/ayaka/internal/domain/shared/nulldatatype"

type GetMaterial struct {
	Number       uint                      `json:"number"`
	DONumber     string                    `db:"DocNoMaterialTransfer" json:"document_number_material"`
	Date         string                    `db:"DocDt" json:"date"`
	ItCode       string                    `db:"ItCode" json:"item_code"`
	ItName       string                    `db:"ItName" json:"item_name"`
	BatchNo      string                    `db:"BatchNo" json:"batch"`
	QtyRemaining float32                   `db:"QtyRemaining" json:"qty_transfer"`
	Uom          string                    `db:"UomName" json:"uom_name"`
	Remark       nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Read struct {
	Number  uint                      `json:"number"`
	DocNo   string                    `db:"DocNo" json:"document_number"`
	Date    string                    `db:"DocDt" json:"date"`
	WhsFrom string                    `db:"WhsFrom" json:"warehouse_from"`
	WhsTo   string                    `db:"WhsTo" json:"warehouse_to"`
	ItName  string                    `db:"ItName" json:"item_name"`
	BatchNo string                    `db:"BatchNo" json:"batch"`
	Qty     float32                   `db:"Qty" json:"quantity"`
	UomName string                    `db:"UomName" json:"uom_name"`
	Remark  nulldatatype.NullDataType `db:"Remark" json:"remark"`
}
