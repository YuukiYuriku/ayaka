package tbldirectsalesdelivery

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo   string                    `db:"DocNo" json:"document_number"`
	DNo     string                    `db:"DNo" json:"detail_number"`
	Cancel  booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	ItCode  string                    `db:"ItCode" json:"item_code"`
	ItName  string                    `db:"ItName" json:"item_name"`
	UomName string                    `db:"UomName" json:"uom_name"`
	BatchNo string                    `db:"BatchNo" json:"batch"`
	Source  string                    `db:"Source" json:"source"`
	Stock   float32                   `db:"Stock" json:"stock"`
	Qty     float32                   `db:"Qty" json:"quantity"`
	Price   float32                   `db:"Price" json:"price"`
}

type Read struct {
	Number        uint                      `json:"number"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"document_date"`
	WhsCode       string                    `db:"WhsCode" json:"warehouse_code"`
	WhsName       string                    `db:"WhsName" json:"warehouse_name"`
	Customer      string                    `db:"CustomerName" json:"customer_name"`
	Address       nulldatatype.NullDataType `db:"Address" json:"address"`
	CityCode      nulldatatype.NullDataType `db:"CityCode" json:"city_code"`
	PostalCode    nulldatatype.NullDataType `db:"PostalCode" json:"postal_code"`
	Phone         nulldatatype.NullDataType `db:"Phone" json:"phone"`
	Email         nulldatatype.NullDataType `db:"Email" json:"email"`
	Mobile        nulldatatype.NullDataType `db:"Mobile" json:"mobile"`
	TaxCode       nulldatatype.NullDataType `db:"TaxCode" json:"tax_code"`
	TaxRate       float32                   `db:"TaxRate" json:"tax_rate"`
	TotalAmount   float32                   `json:"total_amount"`
	TotalQuantity float32                   `json:"total_quantity"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	TblDate       string                    `json:"table_date"`
	Details       []Detail                  `db:"Detail" json:"details"`
}

type Create struct {
	DocNo      string                    `db:"DocNo" json:"document_number" validate:"incolumn=tbldirectsalesdelivhdr->DocNo"`
	Date       string                    `db:"DocDt" json:"document_date" validate:"required"`
	WhsCode    string                    `db:"WhsCode" json:"warehouse_code" validate:"required"`
	WhsName    string                    `db:"WhsName" json:"warehouse_name"`
	Customer   string                    `db:"Customer" json:"customer_name" validate:"required"`
	Address    nulldatatype.NullDataType `db:"Address" json:"address"`
	CityCode   nulldatatype.NullDataType `db:"CityCode" json:"city_code"`
	PostalCode nulldatatype.NullDataType `db:"PostalCode" json:"postal_code"`
	Phone      nulldatatype.NullDataType `db:"Phone" json:"phone"`
	Email      nulldatatype.NullDataType `db:"Email" json:"email"`
	Mobile     nulldatatype.NullDataType `db:"Mobile" json:"mobile"`
	TaxCode    nulldatatype.NullDataType `db:"TaxCode" json:"tax_code"`
	TaxRate    float32                   `db:"TaxRate" json:"tax_rate"`
	Remark     nulldatatype.NullDataType `db:"Remark" json:"remark"`
	CreateBy   string
	CreateDt   string
	Details    []Detail `db:"Detail" json:"details"`
}
