package tbltax

type Read struct {
	Number       uint    `json:"number"`
	TaxCode      string  `db:"TaxCode" json:"tax_code"`
	TaxName      string  `db:"TaxName" json:"tax_name"`
	TaxRate      float32 `db:"TaxRate" json:"tax_rate"`
	TaxGroupCode string  `db:"TaxGroupCode" json:"tax_group_code"`
	TaxGroupName string  `db:"TaxGroupName" json:"tax_group_name"`
	CreateDate   string  `db:"CreateDt" json:"create_date"`
}

type Create struct {
	TaxCode      string  `db:"TaxCode" json:"tax_code" validate:"required,unique=tbltax->TaxCode,max=12"`
	TaxName      string  `db:"TaxName" json:"tax_name" validate:"required"`
	TaxRate      float32 `db:"TaxRate" json:"tax_rate" validate:"required"`
	TaxGroupCode string  `db:"TaxGroupCode" json:"tax_group_code" validate:"required,incolumn=tbltaxgroup->TaxGroupCode"`
	CreateDate   string  `db:"CreateDt" json:"create_date"`
	CreateBy     string  `db:"CreateBy" json:"create_by"`
}

type Update struct {
	TaxCode        string  `db:"TaxCode" json:"tax_code" validate:"required,incolumn=tbltax->TaxCode,max=12"`
	TaxName        string  `db:"TaxName" json:"tax_name"`
	TaxRate        float32 `db:"TaxRate" json:"tax_rate"`
	TaxGroupCode   string  `db:"TaxGroupCode" json:"tax_group_code" validate:"required,incolumn=tbltaxgroup->TaxGroupCode"`
	LastUpdateDate string  `db:"LastUpDt"`
	LastUpdateBy   string  `db:"LastUpBy"`
}
