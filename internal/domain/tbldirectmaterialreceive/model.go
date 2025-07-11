package tbldirectmaterialreceive

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo       string                    `db:"DocNo" json:"document_number"`
	DNo         string                    `db:"DNo" json:"detail_number"`
	Cancel      booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	ItCode      string                    `db:"ItCode" json:"item_code" validate:"required"`
	ItName      string                    `db:"ItName" json:"item_name"`
	Source      string                    `db:"Source" json:"source"`
	BatchNo     string                    `db:"BatchNo" json:"batch"`
	SenderStock float32                   `db:"SenderStock" json:"sender_stock"`
	Qty         float32                   `db:"Qty" json:"quantity" validate:"min=0"`
	UomName     string                    `db:"UomName" json:"uom_name"`
	Remark      nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Read struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"date"`
	TblDate       string                    `json:"table_date"`
	WhsCodeFrom   string                    `db:"WhsCodeFrom" json:"warehouse_code_from"`
	WhsNameFrom   string                    `db:"WhsNameFrom" json:"warehouse_name_from"`
	WhsCodeTo     string                    `db:"WhsCodeTo" json:"warehouse_code_to"`
	WhsNameTo     string                    `db:"WhsNameTo" json:"warehouse_name_to"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	TotalQuantity float32                   `json:"total_quantity"`
	Details       []Detail                  `json:"details"`
}

type Create struct {
	DocNo       string                    `db:"DocNo" json:"document_number"`
	Date        string                    `db:"DocDt" json:"date" validate:"required"`
	WhsCodeFrom string                    `db:"WhsCodeFrom" json:"warehouse_code_from" validate:"required"`
	WhsNameFrom string                    `db:"WhsNameFrom" json:"warehouse_name_from"`
	WhsCodeTo   string                    `db:"WhsCodeTo" json:"warehouse_code_to" validate:"required"`
	WhsNameTo   string                    `db:"WhsNameTo" json:"warehouse_name_to"`
	Remark      nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy    string                    `json:"create_by"`
	CreateDt    string                    `json:"create_date"`
	Details     []Detail                  `json:"details" validate:"dive"`
}
