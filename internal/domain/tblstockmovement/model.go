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
	DNo              string                    `db:"DNo" json:"d_no"`
	CancelInd        string                    `db:"CancelInd" json:"cancel_ind"`
	Source           string                    `db:"Source" json:"source"`
	Source2          nulldatatype.NullDataType `db:"Source2" json:"source2"`
	DocDt            string                    `db:"DocDt" json:"doc_date"`
	WhsCode          string                    `db:"WhsCode" json:"warehouse_code"`
	WhsName          string                    `db:"WhsName" json:"warehouse_name"`
	Lot              nulldatatype.NullDataType `db:"Lot" json:"lot"`
	Bin              nulldatatype.NullDataType `db:"Bin" json:"bin"`
	ItCode           string                    `db:"ItCode" json:"item_code"`
	ItName           string                    `db:"ItName" json:"item_name"`
	Specification    nulldatatype.NullDataType `db:"Specification" json:"item_specification"`
	UomName          string                    `db:"UomName" json:"uom_name"`
	PropCode         string                    `db:"PropCode" json:"property_code"`
	BatchNo          string                    `db:"BatchNo" json:"batch_no"`
	Qty              float32                   `db:"Qty" json:"quantity"`
	Qty2             float32                   `db:"Qty2" json:"quantity2"`
	Qty3             float32                   `db:"Qty3" json:"quantity3"`
	MovingAvgCurCode nulldatatype.NullDataType `db:"MovingAvgCurCode" json:"moving_avg_currency_code"`
	MovingAvgPrice   float32                   `db:"MovingAvgPrice" json:"moving_avg_price"`
	Remark           nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy         string                    `db:"CreateBy" json:"created_by"`
	CreateDt         string                    `db:"CreateDt" json:"created_date"`
	LastUpdBy        nulldatatype.NullDataType `db:"LastUpBy" json:"last_updated_by"`
	LastUpdDt        nulldatatype.NullDataType `db:"LastUpDt" json:"last_updated_date"`
}
