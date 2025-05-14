package tblcurrency

type Fetch struct {
	Number       uint   `json:"number"`
	CurrencyCode string `db:"CurCode" json:"currency_code"`
	CurrencyName string `db:"CurName" json:"currency_name"`
	CreateDate   string `db:"CreateDt" json:"create_date"`
}

type Create struct {
	CurrencyCode string `db:"CurCode" json:"currency_code" validate:"required,unique=tblcurrency->CurCode,max=3" label:"Currency Code"`
	CurrencyName string `db:"CurName" json:"currency_name" validate:"required,max=40" label:"Currency Name"`
	CreateDate string `db:"CreateDt" json:"create_date"`
	CreateBy string `db:"CreateBy" json:"create_by"`
}

type Update struct {
	CurrencyCode string `db:"CurCode" json:"currency_code" validate:"required,max=3,incolumn=tblcurrency->CurCode" label:"Currency Code"`
	CurrencyName string `db:"CurName" json:"currency_name" validate:"required,max=40" label:"Currency Name"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date"`
	UserCode string `db:"LastUpBy" json:"last_update_by"`
}