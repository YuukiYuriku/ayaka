package tblhistoryofstock

type Read struct {
	Number   uint   `json:"number"`
	ItemCode string `db:"ItCode" json:"item_code"`
	ItemName string `db:"ItName" json:"item_name"`
	Batch    string `db:"BatchNo" json:"batch"`
	Source   string `db:"Source" json:"source"`
}
