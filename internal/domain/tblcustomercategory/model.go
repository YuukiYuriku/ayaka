package tblcustomercategory

type Read struct {
	Number               uint   `json:"number"`
	CustomerCategoryCode string `db:"CustCatCode" json:"customer_category_code"`
	CustomerCategoryName string `db:"CustCatName" json:"customer_category_name"`
	CreateDate           string `db:"CreateDt" json:"create_date"`
}

type Create struct {
	CustomerCategoryCode string `db:"CustCatCode" validate:"required,unique=tblcustomercategory->CustCatCode,max=50" json:"customer_category_code"`
	CustomerCategoryName string `db:"CustCatName" validate:"required" json:"customer_category_name"`
	CreateDate           string `db:"CreateDt" json:"create_date"`
	CreateBy             string `db:"CreateBy" json:"create_by"`
}

type Update struct {
	CustomerCategoryCode string `db:"CustCatCode" validate:"required,incolumn=tblcustomercategory->CustCatCode,max=50" json:"customer_category_code"`
	CustomerCategoryName string `db:"CustCatName" validate:"required" json:"customer_category_name"`
	LastUpdateDate       string `db:"CreateDt" json:"last_update_date"`
	LastUpdateBy         string `db:"CreateBy" json:"last_update_by"`
}
