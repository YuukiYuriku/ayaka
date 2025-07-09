package application

import (
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/adapter/rest"
	"gitlab.com/ayaka/internal/application/api"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
)

type Api struct {
	App                           *rest.Fiber                         `inject:"fiber"`
	Config                        *config.Config                      `inject:"config"`
	HealthCheckHandler            api.HealthCheckAPI                  `inject:"healthCheckHandler"`
	TblUserHandler                api.TblUserApi                      `inject:"tblUserHandler"`
	TblCountryHandler             api.TblCountryApi                   `inject:"tblCountryHandler"`
	TblCityHandler                api.TblCityApi                      `inject:"tblCityHandler"`
	TblProvinceHandler            api.TblProvinceAPI                  `inject:"tblProvinceHandler"`
	MiddlewareHandler             *custommiddleware.MiddlewareHandler `inject:"middlewareHandler"`
	TblUomHandler                 api.TblUomApi                       `inject:"tblUomHandler"`
	TblCoaHandler                 api.TblCoaApi                       `inject:"tblCoaHandler"`
	TblItemCatHandler             api.TblItemCatApi                   `inject:"tblItemCatHandler"`
	TblWarehouseHandler           api.TblWarehouseApi                 `inject:"tblWarehouseHandler"`
	TblWarehouseCategoryHandler   api.TblWarehouseCategoryApi         `inject:"tblWarehouseCategoryHandler"`
	TblLogHandler                 api.TblLogApi                       `inject:"tblLogHandler"`
	TblItemHandler                api.TblItemApi                      `inject:"tblItemHandler"`
	TblCurrencyHandler            api.TblCurrencyApi                  `inject:"tblCurrencyHandler"`
	TblInitStockHandler           api.TblInitStockApi                 `inject:"tblInitStockHandler"`
	TblStockAdjustHandler         api.TblStockAdjustApi               `inject:"tblStockAdjustHandler"`
	TblStockSummaryHandler        api.TblStockSummaryApi              `inject:"tblStockSummaryHandler"`
	TblStockMutationHandler       api.TblStockMutationApi             `inject:"tblStockMutationHandler"`
	TblStockMovement              api.TblStockMovementApi             `inject:"tblStockMovementHandler"`
	TblDirectPurchaseReceive      api.TblDirectPurchaseReceiveApi     `inject:"tblDirectPurchaseReceiveHandler"`
	TblTaxGroupHandler            api.TblTaxGroupApi                  `inject:"tblTaxGroupHandler"`
	TblTaxHandler                 api.TblTaxApi                       `inject:"tblTaxHandler"`
	TblCustomerCategoryHandler    api.TblCustomerCategoryApi          `inject:"tblCustomerCategoryHandler"`
	TblSiteHandler                api.TblSiteApi                      `inject:"tblSiteHandler"`
	TblVendorCategoryHandler      api.TblVendorCategoryApi            `inject:"tblVendorCategoryHandler"`
	TblVendorRatingHandler        api.TblVendorRatingApi              `inject:"tblVendorRatingHandler"`
	TblVendorSectorHandler        api.TblVendorSectorApi              `inject:"tblVendorSectorHandler"`
	TblVendorHandler              api.TblVendorApi                    `inject:"tblVendorHandler"`
	TblHistoryOfStockHandler      api.TblHistoryOfStockApi            `inject:"tblHistoryOfStockHandler"`
	TblDailyStockMovementHandler  api.TblDailyStockMovementApi        `inject:"tblDailyStockMovementHandler"`
	TblDirectPurchaseRcvHandler   api.TblDirectPurchaseRcvApi         `inject:"tblDirectPurchaseRcvHandler"`
	TblDirectSalesDeliveryHandler api.TblDirectSalesDeliveryApi       `inject:"tblDirectSalesDeliveryHandler"`
}

func (a *Api) Startup() error {

	a.App.Get("/ping", a.HealthCheckHandler.Ping)
	a.App.Get("/ready", a.HealthCheckHandler.Ready)

	// login endpoint
	a.App.Post("/login", a.TblUserHandler.Login)

	// version
	a.App.Get("/version", a.HealthCheckHandler.PrintVersion)

	// send email
	a.App.Post("/forgot-password", a.TblUserHandler.SendEmailForgotPassword)
	// reset password
	a.App.Put("/reset-password/:token", a.TblUserHandler.ChangePassword)

	logout := a.App.Group("/logout")
	logout.Use(a.MiddlewareHandler.AuthRequired())
	logout.Get("/", a.TblUserHandler.Logout)

	// version v1 group
	v1 := a.App.Group("/v1")
	v1.Use(a.MiddlewareHandler.AuthRequired())

	// log route
	log := v1.Group("/log")
	log.Get("/", a.TblLogHandler.GetLog)

	// country routes group
	country := v1.Group("/country")
	country.Get("/", a.TblCountryHandler.FetchCountries) //get and search countries with pagination and all countries without pagination
	country.Post("/", a.TblCountryHandler.Create)        //create a new country
	country.Put("/:code", a.TblCountryHandler.Update)    //edit a country by country code

	province := v1.Group("/province")
	province.Get("/", a.TblProvinceHandler.FetchProvinces)         // Get and search provinces with pagination
	province.Get("/group", a.TblProvinceHandler.GetGroupProvinces) // Get all provinces grouped by criteria
	province.Get("/:search", a.TblProvinceHandler.DetailProvince)  // Get detail of a province by province code
	province.Post("/", a.TblProvinceHandler.Create)                // Create a new province
	province.Put("/:id", a.TblProvinceHandler.Update)              // Update a province by province code

	//city routes group
	city := v1.Group("/city")
	city.Get("/group", a.TblCityHandler.GetGroupCities) //get all cities group by province
	city.Get("/", a.TblCityHandler.FetchCities)         //get and search cities with pagination
	city.Post("/", a.TblCityHandler.Create)             //create a new city
	city.Put("/:code", a.TblCityHandler.Update)         //update a city by city code

	// UOM routes group
	uom := v1.Group("/uom")
	uom.Get("/", a.TblUomHandler.FetchUom) //get and search uoms with pagination and all uoms data without pagination
	uom.Post("/", a.TblUomHandler.Create)  //create a new uom
	uom.Put("/:code", a.TblUomHandler.Update)

	coa := v1.Group("/coa")
	coa.Get("/", a.TblCoaHandler.FetchCoa) //get and search co

	itemCat := v1.Group("/item-category")
	itemCat.Get("/", a.TblItemCatHandler.FetchItemCategories)
	itemCat.Post("/", a.TblItemCatHandler.Create)
	itemCat.Put("/:code", a.TblItemCatHandler.Update)

	// Warehouse routes group
	warehouse := v1.Group("/warehouse")
	warehouse.Get("/", a.TblWarehouseHandler.FetchWarehouse)         // Get and search warehouses with pagination
	warehouse.Get("/:search", a.TblWarehouseHandler.DetailWarehouse) // Get detail of a warehouse by warehouse code
	warehouse.Post("/", a.TblWarehouseHandler.Create)                // Create a new warehouse
	warehouse.Put("/:code", a.TblWarehouseHandler.Update)            // Update a warehouse by warehouse code

	// Warehouse Category routes group
	warehouseCategory := v1.Group("/warehouse-category")
	warehouseCategory.Get("/", a.TblWarehouseCategoryHandler.FetchWarehouseCategory)
	warehouseCategory.Get("/:search", a.TblWarehouseCategoryHandler.DetailWarehouseCategory)
	warehouseCategory.Post("/", a.TblWarehouseCategoryHandler.Create)
	warehouseCategory.Put("/:code", a.TblWarehouseCategoryHandler.Update)

	item := v1.Group("/item")
	item.Get("/", a.TblItemHandler.FetchItems)    //get and search items with pagination and all
	item.Get("/:search", a.TblItemHandler.Detail) //get details from a spesific item
	item.Post("/", a.TblItemHandler.Create)       //create an item
	item.Put("/:code", a.TblItemHandler.Update)   //update an item

	currency := v1.Group("/currency")
	currency.Get("/", a.TblCurrencyHandler.Fetch)       //get and search currency
	currency.Post("/", a.TblCurrencyHandler.Create)     // create a new currency
	currency.Put("/:code", a.TblCurrencyHandler.Update) // update a currency

	initStock := v1.Group("/initial-stock")
	initStock.Get("/", a.TblInitStockHandler.Fetch)       // get and search initial stock
	initStock.Get("/:code", a.TblInitStockHandler.Detail) // get detail initial stock
	initStock.Post("/", a.TblInitStockHandler.Create)     // Create a new initial stock
	initStock.Put("/:code", a.TblInitStockHandler.Update) // update a currency

	stockAdjust := v1.Group("/stock-adjustment")
	stockAdjust.Get("/", a.TblStockAdjustHandler.Fetch)       //  get and search stock adjjustment
	stockAdjust.Get("/:code", a.TblStockAdjustHandler.Detail) // get a spesific deteail
	stockAdjust.Post("/", a.TblStockAdjustHandler.Create)     // create a new stock adjustment

	// stock summary
	stockSummary := v1.Group("stock-summary")
	stockSummary.Get("/", a.TblStockSummaryHandler.Fetch) //get reporting stock summary

	getItem := v1.Group("get-item")
	getItem.Get("/", a.TblStockSummaryHandler.GetItem) // get all item in a warehouse

	// stock mutation
	stockMutation := v1.Group("stock-mutation")
	stockMutation.Get("/", a.TblStockMutationHandler.Fetch)
	stockMutation.Get("/:code", a.TblStockMutationHandler.Detail)
	stockMutation.Post("/", a.TblStockMutationHandler.Create)
	stockMutation.Put("/:code", a.TblStockMutationHandler.Update)

	// stock movement
	stockMovement := v1.Group("stock-movement")
	stockMovement.Get("/", a.TblStockMovement.Fetch) // get stock movement

	// direct purchase receive
	// directPurchaseReceive := v1.Group("direct-purchase-receive")
	// directPurchaseReceive.Get("/", a.TblDirectPurchaseReceive.Fetch)
	// directPurchaseReceive.Get("/:code", a.TblDirectPurchaseReceive.Detail)
	// directPurchaseReceive.Post("/", a.TblDirectPurchaseReceive.Create)
	// directPurchaseReceive.Put("/:code", a.TblDirectPurchaseReceive.Update)

	// tax group
	taxGroup := v1.Group("/tax-group")
	taxGroup.Get("/", a.TblTaxGroupHandler.Fetch)
	taxGroup.Post("/", a.TblTaxGroupHandler.Create)
	taxGroup.Put("/:code", a.TblTaxGroupHandler.Update)

	// tax
	tax := v1.Group("/tax")
	tax.Get("/", a.TblTaxHandler.Fetch)
	tax.Post("/", a.TblTaxHandler.Create)
	tax.Put("/:code", a.TblTaxHandler.Update)

	// customer category
	customerCategory := v1.Group("/customer-category")
	customerCategory.Get("/", a.TblCustomerCategoryHandler.Fetch)
	customerCategory.Post("/", a.TblCustomerCategoryHandler.Create)
	customerCategory.Put("/:code", a.TblCustomerCategoryHandler.Update)

	// site
	site := v1.Group("/site")
	site.Get("/", a.TblSiteHandler.Fetch)
	site.Post("/", a.TblSiteHandler.Create)
	site.Put("/:code", a.TblSiteHandler.Update)

	// vendor category
	vendorCat := v1.Group("/vendor-category")
	vendorCat.Get("/", a.TblVendorCategoryHandler.Fetch)
	vendorCat.Post("/", a.TblVendorCategoryHandler.Create)
	vendorCat.Put("/:code", a.TblVendorCategoryHandler.Update)

	// vendor rating
	vendorRating := v1.Group("/vendor-rating")
	vendorRating.Get("/", a.TblVendorRatingHandler.Fetch)
	vendorRating.Post("/", a.TblVendorRatingHandler.Create)
	vendorRating.Put("/:code", a.TblVendorRatingHandler.Update)

	// vendor sector
	vendorSector := v1.Group("/vendor-sector")
	vendorSector.Get("/", a.TblVendorSectorHandler.Fetch)
	vendorSector.Post("/", a.TblVendorSectorHandler.Create)
	vendorSector.Put("/:code", a.TblVendorSectorHandler.Update)

	getSector := vendorSector.Group("/get-sector")
	getSector.Get("/", a.TblVendorSectorHandler.GetSector)
	getSector.Get("/:code", a.TblVendorSectorHandler.GetSubSector)

	// master vendor
	vendor := v1.Group("/master-vendor")
	vendor.Get("/", a.TblVendorHandler.Fetch)
	vendor.Get("/:code", a.TblVendorHandler.Detail)
	vendor.Post("/", a.TblVendorHandler.Create)
	vendor.Put("/:code", a.TblVendorHandler.Update)

	// history of stock
	historyOfStock := v1.Group("/history-of-stock")
	historyOfStock.Get("/", a.TblHistoryOfStockHandler.Fetch)

	// daily stock movement
	dailyStockMovement := v1.Group("/daily-stock-movement")
	dailyStockMovement.Get("/", a.TblDailyStockMovementHandler.Fetch)

	// direct purchase receive
	directPurchaseRcv := v1.Group("/direct-purchase-receive")
	directPurchaseRcv.Get("/", a.TblDirectPurchaseRcvHandler.Fetch)
	directPurchaseRcv.Post("/", a.TblDirectPurchaseRcvHandler.Create)
	directPurchaseRcv.Put("/:code", a.TblDirectPurchaseRcvHandler.Update)

	// direct sales delivery
	directSalesDelivery := v1.Group("/direct-sales-delivery")
	directSalesDelivery.Get("/", a.TblDirectSalesDeliveryHandler.Fetch)
	directSalesDelivery.Post("/", a.TblDirectSalesDeliveryHandler.Create)
	directSalesDelivery.Put("/:code", a.TblDirectSalesDeliveryHandler.Update)

	return nil
}

func (a *Api) Shutdown() error {
	return nil
}
