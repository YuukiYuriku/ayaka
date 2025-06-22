package tblvendorcategory

type Read struct {
	Number             uint   `json:"number"`
	VendorCategoryCode string `db:"VendorCatCode" json:"vendor_category_code"`
	VendorCategoryName string `db:"VendorCatName" json:"vendor_category_name"`
	CreateDate         string `db:"CreateDt" json:"create_date"`
}

type Create struct {
	VendorCategoryCode string `json:"vendor_category_code" validate:"required,unique=tblvendorcategory->vendorcatcode,whitespace,max=50"`
	VendorCategoryName string `json:"vendor_category_name" validate:"required,max=50"`
	CreateDate         string `json:"create_date"`
	CreateBy           string `json:"create_by"`
}

type Update struct {
	VendorCategoryCode string `json:"vendor_category_code" validate:"required,incolumn=tblvendorcategory->vendorcatcode,whitespace,max=50"`
	VendorCategoryName string `json:"vendor_category_name" validate:"required,max=50"`
	LastUpdateDate     string `json:"last_update_date"`
	LastUpdateBy       string `json:"last_update_by"`
}
