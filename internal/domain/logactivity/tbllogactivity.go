package logactivity

import "gitlab.com/ayaka/internal/pkg/customerrors"

type LogActivity struct {
	Log  string `db:"Log" json:"log"`
	Date string `db:"Date" json:"date"`
}

var listTable = map[string]string{
	"Country":            "tblcountry",
	"Province":           "tblprovince",
	"City":               "tblcity",
	"Uom":                "tbluom",
	"ItemCategory":       "tblitemcategory",
	"Item":               "tblitem",
	"Warehouse":          "tblwarehouse",
	"WarehouseCategory":  "tblwarehousecategory",
	"Currency":           "tblcurrency",
	"StockInitial":       "tblstockinitialhdr",
	"StockAdjustment":    "tblstockadjustmenthdr",
	"StockMutation":      "tblstockmutationhdr",
	"StockInitialDtl":    "tblstockinitialdtl",
	"StockAdjustmentDtl": "tblstockadjustmentdtl",
	"StockMutationDtl":   "tblstockmutationdtl",
	// "DirectPurchaseReceive":    "tblrecvvdhdr",
	// "DirectPurchaseReceiveDtl": "tblrecvvddtl",
	"TaxGroup":              "tbltaxgroup",
	"Tax":                   "tbltax",
	"CustomerCategory":      "tblcustomercategory",
	"Site":                  "tblsite",
	"VendorCategory":        "tblvendorcategory",
	"VendorRating":          "tblvendorrating",
	"VendorSector":          "tblvendorsectorhdr",
	"MasterVendor":          "tblvendorhdr",
	"DirectPurchaseReceive": "tbldirectpurchasercvhdr",
	"DirectSalesDelivery":   "tbldirectsalesdelivhdr",
	"MaterialTransfer":      "tblmaterialtransferhdr",
	"MaterialReceive":       "tblmaterialreceivehdr",
	"DirectMaterialReceive": "tbldirectmaterialreceivehdr",
}

var listCode = map[string]string{
	"Country":               "CntCode",
	"Province":              "ProvCode",
	"City":                  "CityCode",
	"Uom":                   "UomCode",
	"ItemCategory":          "ItCtCode",
	"Item":                  "ItCode",
	"Warehouse":             "WhsCode",
	"WarehouseCategory":     "WhsCtCode",
	"Currency":              "CurCode",
	"StockInitial":          "DocNo",
	"StockAdjustment":       "DocNo",
	"StockMutation":         "DocNo",
	"StockInitialDtl":       "DNo",
	"StockAdjustmentDtl":    "DNo",
	"StockMutationDtl":      "DNo",
	"DirectPurchaseReceive": "DocNo",
	// "DirectPurchaseReceiveDtl": "DNo",
	"TaxGroup":              "TaxGroupCode",
	"Tax":                   "TaxCode",
	"CustomerCategory":      "CustCatCode",
	"Site":                  "SiteCode",
	"VendorCategory":        "VendorCatCode",
	"VendorRating":          "IndicatorCode",
	"VendorSector":          "SectorCode",
	"MasterVendor":          "VendorCode",
	"DirectSalesDelivery":   "DocNo",
	"MaterialTransfer":      "DocNo",
	"MaterialReceive":       "DocNo",
	"DirectMaterialReceive": "DocNo",
}

var listDoc = map[string]string{
	"StockInitial":          "SI",
	"StockAdjustment":       "SA",
	"StockMutation":         "SM",
	"DirectPurchaseReceive": "DPRV",
	"DirectSalesDelivery":   "DSDV",
	"MaterialTransfer":      "MTF",
	"MaterialReceive":       "MRV",
	"DirectMaterialReceive": "DMRV",
}

func TableOf(category string) (string, error) {
	if table, exists := listTable[category]; exists {
		return table, nil
	}
	return "", customerrors.ErrKeyNotFound
}

func PrimaryKeyOf(category string) (string, error) {
	if code, exists := listCode[category]; exists {
		return code, nil
	}
	return "", customerrors.ErrKeyNotFound
}

func DocNumberOf(category string) (string, error) {
	if code, exists := listDoc[category]; exists {
		return code, nil
	}
	return "", customerrors.ErrKeyNotFound
}
