package tblorderreport

type OrderReportByVendor struct {
	Number                uint    `json:"number"`
	VendorName            string  `db:"VendorName" json:"vendor_name"`
	AveragePrice          float32 `db:"AveragePrice" json:"average_price"`
	TotalOrderFrequency   float32 `db:"OrderFreq" json:"total_order_frequency"`
	TotalOrderAmount      float32 `db:"TotalOrderAmount" json:"total_order_amount"`
	OrderAmountPercentage float32 `db:"TotalOrderPercent" json:"total_order_percentage"`
}
