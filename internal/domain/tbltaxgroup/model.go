package tbltaxgroup

type ReadTblTaxGroup struct {
	Number       uint   `json:"number"`
	TaxGroupCode string `db:"TaxGroupCode" json:"tax_group_code"`
	TaxGroupName string `db:"TaxGroupName" json:"tax_group_name"`
	CreateDate   string `db:"CreateDt" json:"create_date"`
}

type Create struct {
	TaxGroupCode string `json:"tax_group_code" validate:"required,max=12,unique=tbltaxgroup->TaxGroupCode"`
	TaxGroupName string `json:"tax_group_name" validate:"required,max=255"`
	CreateDate   string `db:"CreateDt" json:"create_date"`
	CreateBy     string `db:"CreateBy" json:"create_by"`
}

type Update struct {
	TaxGroupCode string `json:"tax_group_code" validate:"required,max=12,incolumn=tbltaxgroup->TaxGroupCode"`
	TaxGroupName string `json:"tax_group_name" validate:"required"`
	LastUpdateDate   string `db:"LastUpDt" json:"last_update_date"`
	LastUpdateBy     string `db:"LastUpBy" json:"last_update_by"`
}