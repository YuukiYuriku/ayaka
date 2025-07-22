package dashboard

type Read struct {
	OutstandingMaterialTransfer uint `db:"OutstandingMaterialTransfer" json:"outstanding_material_transfer"`
	OutstandingPurchaseOrder    uint `db:"OutstandingPurchaseOrder" json:"outstanding_purchase_material_order"`
	OutstandingMaterialOrder    uint `db:"OutstandingMaterialOrder" json:"outstanding_material_order"`
}