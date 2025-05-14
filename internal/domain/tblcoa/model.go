package tblcoa

type ReadTblCoa struct {
	Number uint `json:"number"`
	AccountNumber string `db:"AcNo" json:"account"`
	Description string `db:"AcDesc" json:"description"`
}