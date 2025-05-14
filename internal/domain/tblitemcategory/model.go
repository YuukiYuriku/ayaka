package tblitemcategory

import (
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
)

type ReadTblItemCategory struct {
	Number                 uint                      `json:"number"`
	ItemCategoryCode       string                    `db:"ItCtCode" json:"item_category_code"`
	ItemCategoryName       string                    `db:"ItCtName" json:"item_category_name"`
	Active                 booldatatype.BoolDataType `db:"ActInd" json:"active"`
	CreateDate             string                    `db:"CreateDt" json:"create_date"`
	CoaStock               nulldatatype.NullDataType `db:"AcNo" json:"coa_stock"`
	CoaStockDesc           nulldatatype.NullDataType `db:"AcDesc" json:"coa_stock_description"`
	CoaSales               nulldatatype.NullDataType `db:"AcNo2" json:"coa_sales"`
	CoaSalesDesc           nulldatatype.NullDataType `db:"AcDesc2" json:"coa_sales_description"`
	CoaCOGS                nulldatatype.NullDataType `db:"AcNo3" json:"coa_cogs"`
	CoaCOGSDesc            nulldatatype.NullDataType `db:"AcDesc3" json:"coa_cogs_description"`
	CoaSalesReturn         nulldatatype.NullDataType `db:"AcNo4" json:"coa_sales_return"`
	CoaSalesReturnDesc     nulldatatype.NullDataType `db:"AcDesc4" json:"coa_sales_return_description"`
	CoaPurchaseReturn      nulldatatype.NullDataType `db:"AcNo5" json:"coa_purchase_return"`
	CoaPurchaseReturnDesc  nulldatatype.NullDataType `db:"AcDesc5" json:"coa_purchase_return_description"`
	CoaConsumptionCost     nulldatatype.NullDataType `db:"AcNo6" json:"coa_consumption_cost"`
	CoaConsumptionCostDesc nulldatatype.NullDataType `db:"AcDesc6" json:"coa_consumption_cost_description"`
}

type Create struct {
	ItemCategoryCode   string                    `db:"ItCtCode" json:"item_category_code" validate:"required,max=16,unique=tblitemcategory->ItCtCode" label:"Item Category Code"`
	ItemCategoryName   string                    `db:"ItCtName" json:"item_category_name" validate:"required" label:"Item Category Name"`
	Active             booldatatype.BoolDataType `db:"ActInd" json:"active" validate:"required"`
	CoaStock           nulldatatype.NullDataType `db:"AcNo" json:"coa_stock" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Stock)"`
	CoaSales           nulldatatype.NullDataType `db:"AcNo2" json:"coa_sales" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Sales)"`
	CoaCOGS            nulldatatype.NullDataType `db:"AcNo3" json:"coa_cogs" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(COGS)"`
	CoaSalesReturn     nulldatatype.NullDataType `db:"AcNo4" json:"coa_sales_return" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Sales Return)"`
	CoaPurchaseReturn  nulldatatype.NullDataType `db:"AcNo5" json:"coa_purchase_return" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Purchase Return)"`
	CoaConsumptionCost nulldatatype.NullDataType `db:"AcNo6" json:"coa_consumption_cost" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Consumption Cost)"`
	CreateDate         string                    `db:"CreateDt" json:"create_date"`
	CreateBy           string                    `db:"CreateBy" json:"create_by"`
}

type Update struct {
	ItemCategoryCode   string                    `db:"ItCtCode" json:"item_category_code" validate:"required,max=16,incolumn=tblitemcategory->ItCtCode" label:"Item Category Code"`
	ItemCategoryName   string                    `db:"ItCtName" json:"item_category_name" validate:"required" label:"Item Category Name"`
	Active             booldatatype.BoolDataType `db:"ActInd" json:"active"`
	CoaStock           nulldatatype.NullDataType `db:"AcNo" json:"coa_stock" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Stock)"`
	CoaSales           nulldatatype.NullDataType `db:"AcNo2" json:"coa_sales" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Sales)"`
	CoaCOGS            nulldatatype.NullDataType `db:"AcNo3" json:"coa_cogs" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(COGS)"`
	CoaSalesReturn     nulldatatype.NullDataType `db:"AcNo4" json:"coa_sales_return" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Sales Return)"`
	CoaPurchaseReturn  nulldatatype.NullDataType `db:"AcNo5" json:"coa_purchase_return" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Purchase Return)"`
	CoaConsumptionCost nulldatatype.NullDataType `db:"AcNo6" json:"coa_consumption_cost" validate:"omitempty,incolumn=tblcoa->AcNo" label:"COA Account(Consumption Cost)"`
	LastUpdateDate     string                    `db:"LastUpDt" json:"last_update_date"`
	LastUpdateBy       string                    `db:"LastUpBy" json:"last_update_by"`
}
