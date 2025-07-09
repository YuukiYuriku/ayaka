package tblstockmutationdtl

type Detail struct {
	DocNo string `db:"DocNo" json:"document_number"`
	DNo      string  `db:"DNo" json:"detail_number"`
	ItemCode string  `db:"ItCode" json:"item_code"`
	ItemName string  `db:"ItName" json:"item_name"`
	Batch    string  `db:"BatchNo" json:"batch"`
	Source   string  `db:"Source" json:"source"`
	Stock    float32 `db:"Stock" json:"stock"`
	Quantity float32 `db:"Qty" json:"quantity"`
	Uom      string  `db:"UomName" json:"uom"`
}

type Create struct {
	DocNo   string  `json:"document_number"`
	DNo     string  `json:"d_no"`
	ItCode  string  `json:"item_code" validate:"required" label:"Item"`
	BatchNo string  `json:"batch"`
	Source  string  `json:"source"`
	Stock   float32 `json:"stock"`
	Qty     float32 `json:"quantity" validate:"min=1" label:"Quantity"`
}
