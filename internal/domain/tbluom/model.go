package tbluom

type ReadTblUom struct {
	Number     uint   `json:"number"`
	UomCode    string `db:"UomCode" json:"uom_code"`
	UomName    string `db:"UomName" json:"uom_name"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type CreateTblUom struct {
	UomCode    string `db:"UomCode" json:"uom_code" validate:"required,unique=tbluom->UomCode" label:"Uom Code"`
	UomName    string `db:"UomName" json:"uom_name" validate:"required,unique=tbluom->UomName" label:"Uom Name"`
	CreateBy   string `db:"CreateBy" json:"create_by"`
	CreateDate string `db:"CreateDt" json:"create_date"`
}

type UpdateTblUom struct {
	UomCode        string `db:"UomCode" json:"uom_code" validate:"required,incolumn=tbluom->UomCode" label:"Uom Code"`
	UomName        string `db:"UomName" json:"uom_name" validate:"unique=tbluom->UomName" label:"Uom Name"`
	UserCode       string `db:"UserCode" json:"user_code"`
	LastUpdateDate string `db:"LastUpDt" json:"last_update_date"`
}
