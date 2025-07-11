package tblmaterialtransfer

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Detail struct {
	DocNo     string                    `db:"DocNo" json:"document_number"`
	DNo       string                    `db:"DNo" json:"detail_number"`
	ItCode    string                    `db:"ItCode" json:"item_code" validate:"required"`
	ItName    string                    `db:"ItName" json:"item_name"`
	BatchNo   string                    `db:"BatchNo" json:"batch"`
	Stock     float32                   `db:"Stock" json:"stock"`
	Qty       float32                   `db:"Qty" json:"quantity"`
	UomName   string                    `db:"UomName" json:"uom_name"`
	CancelInd booldatatype.BoolDataType `db:"CancelInd" json:"cancel"`
	SuccesInd booldatatype.BoolDataType `db:"SuccessInd" json:"success"`
	Remark    nulldatatype.NullDataType `db:"Remark" json:"remark"`
}

type Create struct {
	DocNo         string                    `json:"document_number"`
	Date          string                    `json:"date" validate:"required"`
	Status        string                    `json:"status"`
	WhsCodeFrom   string                    `json:"warehouse_code_from" validate:"required,incolumn=tblwarehouse->WhsCode"`
	WhsCodeTo     string                    `json:"warehouse_code_to" validate:"required,incolumn=tblwarehouse->WhsCode"`
	VendorCode    nulldatatype.NullDataType `json:"vendor_code"`
	Driver        nulldatatype.NullDataType `json:"driver"`
	TransportType nulldatatype.NullDataType `json:"transport_type"`
	LicenceNo     nulldatatype.NullDataType `json:"licence"`
	Note          nulldatatype.NullDataType `json:"note"`
	Details       []Detail                  `json:"details" validate:"dive"`
	CreateBy      string                    `json:"create_by"`
	CreateDt      string                    `json:"create_date"`
}

type Read struct {
	Number        uint                      `json:"number"`
	TblDate       string                    `json:"table_date"`
	DocNo         string                    `db:"DocNo" json:"document_number"`
	Date          string                    `db:"DocDt" json:"date" validate:"required"`
	Status        string                    `db:"Status" json:"status"`
	WhsCodeFrom   string                    `db:"WhsCodeFrom" json:"warehouse_code_from" validate:"required"`
	WhsNameFrom   string                    `db:"WhsNameFrom" json:"warehouse_name_from"`
	WhsCodeTo     string                    `db:"WhsCodeTo" json:"warehouse_code_to" validate:"required"`
	WhsNameTo     string                    `db:"WhsNameTo" json:"warehouse_name_to"`
	VendorCode    nulldatatype.NullDataType `db:"VendorCode" json:"vendor_code"`
	Driver        nulldatatype.NullDataType `db:"Driver" json:"driver"`
	TransportType nulldatatype.NullDataType `db:"TransportType" json:"transport_type"`
	LicenceNo     nulldatatype.NullDataType `db:"LicenceNo" json:"licence"`
	Note          nulldatatype.NullDataType `db:"Note" json:"note"`
	Details       []Detail                  `json:"details" validate:"dive"`
}
