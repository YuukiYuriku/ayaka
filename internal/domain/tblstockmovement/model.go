package tblstockmovement

import (
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
	Number           uint                      `json:"number"`
	DocType          string                    `db:"DocType" json:"doc_type"`
	FromTo           nulldatatype.NullDataType `db:"FromTo" json:"from_to,omitempty"`
	DocNo            string                    `db:"DocNo" json:"doc_no"`
	Source           string                    `db:"Source" json:"source"`
	DocDt            string                    `db:"DocDt" json:"doc_date"`
	WhsName          string                    `db:"WhsName" json:"warehouse_name"`
	ItCode           string                    `db:"ItCode" json:"item_code"`
	ItName           string                    `db:"ItName" json:"item_name"`
	Specification    nulldatatype.NullDataType `db:"Specification" json:"item_specification"`
	UomName          string                    `db:"UomName" json:"uom_name"`
	BatchNo          string                    `db:"BatchNo" json:"batch_no"`
	Qty              float32                   `db:"Qty" json:"quantity"`
	Remark           nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy         string                    `db:"CreateBy" json:"created_by"`
	CreateDt         string                    `db:"CreateDt" json:"created_date"`
}
