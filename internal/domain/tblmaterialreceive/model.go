package tblmaterialreceive

import "gitlab.com/ayaka/internal/domain/shared/nulldatatype"

type Detail struct {
	DocNo                 string                    `db:"DocNo" json:"document_number"`
	DNo                   string                    `db:"DNo" json:"detail_number"`
	DocNoMaterialTransfer string                    `db:"DocNoMaterialTransfer" json:"document_number_material"`
	ItCode                string                    `db:"ItCode" json:"item_code"`
	ItName                string                    `db:"ItName" json:"item_name"`
	UomName               string                    `db:"UomName" json:"uom_name"`
	BatchNo               string                    `db:"BatchNo" json:"batch"`
	Source                string                    `db:"Source" json:"source"`
	QtyTransfer           float32                   `db:"QtyTransfer" json:"qty_transfer"`
	QtyActual             float32                   `db:"QtyActual" json:"qty_actual"`
	Remark                nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Create struct {
	DocNo       string                    `json:"document_number"`
	Date        string                    `json:"date" validate:"required"`
	WhsCodeFrom string                    `json:"warehouse_code_from" validate:"required"`
	WhsCodeTo   string                    `json:"warehouse_code_to" validate:"required"`
	Remark      nulldatatype.NullDataType `json:"remark"`
	CreateBy    string                    `json:"create_by"`
	CreateDt    string                    `json:"create_date"`
	Details     []Detail                  `json:"details" validate:"dive"`
}

type Read struct {
	Number      uint                      `json:"number"`
	DocNo       string                    `json:"document_number" db:"DocNo" validate:"incolumn=tblmaterialreceive->DocNo"`
	Date        string                    `json:"date" db:"DocDt"`
	TblDate     string                    `json:"table_date"`
	WhsCodeFrom string                    `json:"warehouse_code_from" db:"WhsCodeFrom"`
	WhsNameFrom string                    `json:"warehouse_name_from" db:"WhsNameFrom"`
	WhsCodeTo   string                    `json:"warehouse_code_to" db:"WhsCodeTo"`
	WhsNameTo   string                    `json:"warehouse_name_to" db:"WhsNameTo"`
	Remark      nulldatatype.NullDataType `json:"remark" db:"Remark"`
	CreateBy    string                    `json:"create_by" db:"CreateBy"`
	CreateDt    string                    `json:"create_date" db:"CreateDt"`
	Details     []Detail                  `json:"details"`
}
