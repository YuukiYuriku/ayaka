package bootstrap

import (
	"gitlab.com/ayaka/internal/application"
	"gitlab.com/ayaka/internal/application/api"
	"gitlab.com/ayaka/internal/application/service"
)

func RegisterService() {
	appContainer.RegisterService("tblUserService", new(service.TblUser))
	appContainer.RegisterService("tblCountryService", new(service.TblCountry))
	appContainer.RegisterService("tblProvinceService", new(service.TblProvince))
	appContainer.RegisterService("tblCityService", new(service.TblCity))
	appContainer.RegisterService("tblUomService", new(service.TblUom))
	appContainer.RegisterService("tblCoaService", new(service.TblCoa))
	appContainer.RegisterService("tblItemCatService", new(service.TblItemCat))
	appContainer.RegisterService("tblWarehouseService", new(service.TblWarehouse))
	appContainer.RegisterService("tblWarehouseCategoryService", new(service.TblWarehouseCategory))
	appContainer.RegisterService("tblLogService", new(service.TblLog))
	appContainer.RegisterService("tblItemService", new(service.TblItem))
	appContainer.RegisterService("tblCurrencyService", new(service.TblCurrency))
	appContainer.RegisterService("tblInitStockService", new(service.TblInitStock))
	appContainer.RegisterService("tblStockAdjustService", new(service.TblStockAdjust))
	appContainer.RegisterService("tblStockSummaryService", new(service.TblStockSummary))
	appContainer.RegisterService("tblStockMutationService", new(service.TblStockMutation))
	appContainer.RegisterService("tblStockMovementService", new(service.TblStockMovement))
	appContainer.RegisterService("tblDirectPurchaseReceiveService", new(service.TblDirectPurchaseReceive))
	appContainer.RegisterService("tblTaxGroupService", new(service.TblTaxGroup))
	appContainer.RegisterService("tblTaxService", new(service.TblTax))
	appContainer.RegisterService("tblCustomerCategoryService", new(service.TblCustomerCategory))
	appContainer.RegisterService("tblSiteService", new(service.TblSite))
	appContainer.RegisterService("tblVendorCategoryService", new(service.TblVendorCategory))
	appContainer.RegisterService("tblVendorRatingService", new(service.TblVendorRating))
	appContainer.RegisterService("tblVendorSectorService", new(service.TblVendorSector))
	appContainer.RegisterService("tblVendorService", new(service.TblVendor))
}

func RegisterApi() {
	// appContainer.RegisterService("healthCheckHandler", new(api.HealthCheckHandler))
	// appContainer.RegisterService("tblUserHandler", new(api.TblUserHandler))
	appContainer.RegisterService("healthCheckHandler", new(api.HealthCheckHandler))
	appContainer.RegisterService("tblUserHandler", new(api.TblUserHandler))
	appContainer.RegisterService("tblCountryHandler", new(api.TblCountryHandler))
	appContainer.RegisterService("tblProvinceHandler", new(api.TblProvinceHandler))
	appContainer.RegisterService("tblCityHandler", new(api.TblCityHandler))
	appContainer.RegisterService("tblUomHandler", new(api.TblUomHandler))
	appContainer.RegisterService("tblCoaHandler", new(api.TblCoaHandler))
	appContainer.RegisterService("tblItemCatHandler", new(api.TblItemCatHandler))
	appContainer.RegisterService("tblWarehouseHandler", new(api.TblWarehouseHandler))
	appContainer.RegisterService("tblWarehouseCategoryHandler", new(api.TblWarehouseCategoryHandler))
	appContainer.RegisterService("tblLogHandler", new(api.TblLogHandler))
	appContainer.RegisterService("tblItemHandler", new(api.TblItemHandler))
	appContainer.RegisterService("tblCurrencyHandler", new(api.TblCurrencyHandler))
	appContainer.RegisterService("tblInitStockHandler", new(api.TblInitStockHandler))
	appContainer.RegisterService("tblStockAdjustHandler", new(api.TblStockAdjustHandler))
	appContainer.RegisterService("tblStockSummaryHandler", new(api.TblStockSummaryHandler))
	appContainer.RegisterService("tblStockMutationHandler", new(api.TblStockMutationHandler))
	appContainer.RegisterService("tblStockMovementHandler", new(api.TblStockMovementHandler))
	appContainer.RegisterService("tblDirectPurchaseReceiveHandler", new(api.TblDirectPurchaseReceiveHandler))
	appContainer.RegisterService("tblTaxGroupHandler", new(api.TblTaxGroupHandler))
	appContainer.RegisterService("tblTaxHandler", new(api.TblTaxHandler))
	appContainer.RegisterService("tblCustomerCategoryHandler", new(api.TblCustomerCategoryHandler))
	appContainer.RegisterService("tblSiteHandler", new(api.TblSiteHandler))
	appContainer.RegisterService("tblVendorCategoryHandler", new(api.TblVendorCategoryHandler))
	appContainer.RegisterService("tblVendorRatingHandler", new(api.TblVendorRatingHandler))
	appContainer.RegisterService("tblVendorSectorHandler", new(api.TblVendorSectorHandler))
	appContainer.RegisterService("tblVendorHandler", new(api.TblVendorHandler))

	appContainer.RegisterService("api", new(application.Api))

}
