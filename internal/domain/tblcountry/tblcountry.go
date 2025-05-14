package tblcountry

type Readtblcountry struct {
	Number      uint   `json:"number"`
	CountryCode string `db:"CntCode" json:"country_code"`
	CountryName string `db:"CntName" json:"country_name"`
	CreateDate  string `db:"CreateDt" json:"create_date"`
}

type Createtblcountry struct {
	CountryCode string `db:"CntCode" json:"country_code" validate:"required,min=2,max=3,unique=tblcountry->CntCode" label:"Country Code"`
	CountryName string `db:"CntName" json:"country_name" validate:"required,min=1,max=40" label:"Country Name"`
	CreateBy    string `db:"CreateBy" json:"create_by"`
	CreateDate  string `db:"CreateDt" json:"create_date"`
}

type Updatetblcountry struct {
	CountryCode    string `db:"CntCode" json:"country_code" validate:"required,min=2,max=3,incolumn=tblcountry->CntCode" label:"Country Code"`
	CountryName    string `db:"CntName" json:"country_name" validate:"required,min=1,max=40" label:"Country Name"`
	UserCode       string `db:"UserCode" json:"user_code"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date"`
}
