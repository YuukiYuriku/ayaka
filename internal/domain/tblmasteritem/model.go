package tblmasteritem

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type Read struct {
	Number        uint                      `json:"number"`
	ItemCode      string                    `db:"ItCode" json:"item_code"`
	ItemName      string                    `db:"ItName" json:"item_name"`
	UomName       string                    `db:"UomName" json:"uom_name"`
	LocalCode     nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code"`
	ForeignName   nulldatatype.NullDataType `db:"ForeignName" json:"foreign_name"`
	OldCode       nulldatatype.NullDataType `db:"ItCodeOld" json:"old_code"`
	CategoryCode  string                    `db:"ItCtCode" json:"category_code"`
	Category      string                    `db:"ItCtName" json:"category_name"`
	Spesification nulldatatype.NullDataType `db:"Specification" json:"spesification"`
	Active        booldatatype.BoolDataType `db:"ActInd" json:"active"`
	ItemRequest   nulldatatype.NullDataType `db:"ItemRequestDocNo" json:"item_request"`
	UomCode       string                    `db:"PurchaseUomCode" json:"uom_code"`
	HSCode        nulldatatype.NullDataType `db:"HSCode" json:"hs_code"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	InventoryItem booldatatype.BoolDataType `db:"InventoryItemInd" json:"inventory_item"`
	SalesItem     booldatatype.BoolDataType `db:"SalesItemInd" json:"sales_item"`
	PurchaseItem  booldatatype.BoolDataType `db:"PurchaseItemInd" json:"purchase_item"`
	ServiceItem   booldatatype.BoolDataType `db:"ServiceItemInd" json:"service_item"`
	TaxLiable     booldatatype.BoolDataType `db:"TaxLiableInd" json:"tax_liable"`
	CreateDate    string                    `db:"CreateDt" json:"create_date"`
}

type Detail struct {
	ItemRequest   nulldatatype.NullDataType `db:"ItemRequestDocNo" json:"item_request"`
	Uom           string                    `db:"UomName" json:"uom_name"`
	UomCode       string                    `db:"PurchaseUomCode" json:"uom_code"`
	HSCode        nulldatatype.NullDataType `db:"HSCode" json:"hs_code"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark"`
	InventoryItem booldatatype.BoolDataType `db:"InventoryItemInd" json:"inventory_item"`
	SalesItem     booldatatype.BoolDataType `db:"SalesItemInd" json:"sales_item"`
	PurchaseItem  booldatatype.BoolDataType `db:"PurchaseItemInd" json:"purchase_item"`
	ServiceItem   booldatatype.BoolDataType `db:"ServiceItemInd" json:"service_item"`
	TaxLiable     booldatatype.BoolDataType `db:"TaxLiableInd" json:"tax_liable"`
}

type Create struct {
	ItemCode      string                    `db:"ItCode" json:"item_code"`
	ItemName      string                    `db:"ItName" json:"item_name" validate:"required,max=250" label:"Item Name"`
	LocalCode     nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code" validate:"max=30" label:"Local Code"`
	ForeignName   nulldatatype.NullDataType `db:"ForeignName" json:"foreign_name" validate:"max=250" label:"Foreign Name"`
	OldCode       nulldatatype.NullDataType `db:"ItCodeOld" json:"old_code" validate:"max=30" label:"Old Code"`
	Category      string                    `db:"ItCtCode" json:"category_code" validate:"required,min=1,incolumn=tblitemcategory->ItCtCode" label:"Category"`
	Spesification nulldatatype.NullDataType `db:"Specification" json:"spesification" validate:"max=400"  label:"Spesification"`
	Active        booldatatype.BoolDataType `db:"ActInd" json:"active"`
	CreateDate    string                    `db:"CreateDt" json:"create_date"`
	ItemRequest   nulldatatype.NullDataType `db:"ItemRequestDocNo" json:"item_request" validate:"max=30"  label:"Item Request"`
	Uom           string                    `db:"PurchaseUomCode" json:"uom_code" validate:"required,incolumn=tbluom->UomCode" label:"Uom"`
	HSCode        nulldatatype.NullDataType `db:"HSCode" json:"hs_code" validate:"max=30" label:"HS Code"`
	Remark        nulldatatype.NullDataType `db:"Remark" json:"remark" validate:"max=1000" label:"Remark"`
	InventoryItem booldatatype.BoolDataType `db:"InventoryItemInd" json:"inventory_item"`
	SalesItem     booldatatype.BoolDataType `db:"SalesItemInd" json:"sales_item"`
	PurchaseItem  booldatatype.BoolDataType `db:"PurchaseItemInd" json:"purchase_item"`
	ServiceItem   booldatatype.BoolDataType `db:"ServiceItemInd" json:"service_item"`
	TaxLiable     booldatatype.BoolDataType `db:"TaxLiableInd" json:"tax_liable"`
	CreateBy      string                    `db:"CreateBy" json:"create_by"`
	Source        nulldatatype.NullDataType `db:"ItScCode" json:"source" validate:"max=30" label:"Source"`
}

type Update struct {
	ItemCode       string                    `db:"ItCode" json:"item_code" validate:"incolumn=tblitem->ItCode" label:"Item Code"`
	ItemName       string                    `db:"ItName" json:"item_name" validate:"required,max=250" label:"Item Name"`
	LocalCode      nulldatatype.NullDataType `db:"ItCodeInternal" json:"local_code" validate:"max=30" label:"Local Code"`
	ForeignName    nulldatatype.NullDataType `db:"ForeignName" json:"foreign_name" validate:"max=250" label:"Foreign Name"`
	OldCode        nulldatatype.NullDataType `db:"ItCodeOld" json:"old_code" validate:"max=30" label:"Old Code"`
	Spesification  nulldatatype.NullDataType `db:"Specification" json:"spesification" validate:"max=400" label:"Spesification"`
	Active         booldatatype.BoolDataType `db:"ActInd" json:"active"`
	HSCode         nulldatatype.NullDataType `db:"HSCode" json:"hs_code" validate:"max=30" label:"HS Code"`
	Remark         nulldatatype.NullDataType `db:"Remark" json:"remark" validate:"max=1000" label:"Remark"`
	InventoryItem  booldatatype.BoolDataType `db:"InventoryItemInd" json:"inventory_item"`
	SalesItem      booldatatype.BoolDataType `db:"SalesItemInd" json:"sales_item"`
	PurchaseItem   booldatatype.BoolDataType `db:"PurchaseItemInd" json:"purchase_item"`
	ServiceItem    booldatatype.BoolDataType `db:"ServiceItemInd" json:"service_item"`
	TaxLiable      booldatatype.BoolDataType `db:"TaxLiableInd" json:"tax_liable"`
	LastUpdateBy   string                    `db:"LastUpBy"`
	LastUpdateDate string                    `db:"LastUpDt"`
}

type Check struct {
	ItemCode string `db:"ItCode" json:"item_code"`
	ItemName string `db:"ItName" json:"item_name"`
}
