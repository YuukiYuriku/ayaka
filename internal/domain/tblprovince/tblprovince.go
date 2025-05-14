package tblprovince

// Structure for reading a list of province data
type ReadTblProvince struct {
	Number     uint   `json:"number"`
	ProvCode   string `db:"ProvCode" json:"province_code"`
	ProvName   string `db:"ProvName" json:"province_name"`
	CntCode    string `db:"CntCode" json:"country_code"`
	CntName    string `db:"CntName" json:"country"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type DetailTblProvince struct {
	ProvCode    string                     `db:"ProvCode" json:"province_code"`
	ProvName    string                     `db:"ProvName" json:"province_name"`
	CntCode     string                     `db:"CntCode" json:"country_code"`
	CntName     string                     `db:"CntName" json:"country"`
	CreateBy    string                     `db:"CreateBy" json:"create_by"`
	CreateDate  string                     `db:"CreateDt" json:"create_date"`
}

type CreateTblProvince struct {
	ProvCode    string `db:"ProvCode" json:"province_code" validate:"required,min=1,max=16,unique=tblprovince->ProvCode"`
	ProvName    string `db:"ProvName" json:"province_name" validate:"required,min=1,max=80"`
	CountryCode string `db:"CntCode" json:"country" validate:"required,incolumn=tblcountry->CntCode"`
	CreateBy    string `db:"CreateBy" json:"create_by"`
	CreateDate  string `db:"CreateDt" json:"create_date"`
}

type UpdateTblProvince struct {
	ProvCode       string `db:"ProvCode" json:"province_code" validate:"required,min=1,max=16,incolumn=tblprovince->ProvCode"`
	ProvName       string `db:"ProvName" json:"province_name" validate:"required,min=1,max=80"`
	CountryCode    string `db:"CntCode" json:"country" validate:"required,incolumn=tblcountry->CntCode"`
	UserCode       string `db:"UserCode" json:"user_code"`
	LastUpBy       string `db:"LastUpBy" json:"last_update_by"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date"`
}

// Structure for grouping provinces by country
type GroupProvinceByCountry struct {
	GroupedData string `db:"GroupedData"`
	CountryName string `db:"CountryName"`
}
