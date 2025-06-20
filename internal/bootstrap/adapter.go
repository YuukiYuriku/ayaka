package bootstrap

import (
	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/adapter/repository/cache"
	"gitlab.com/ayaka/internal/adapter/repository/sqlx"
	"gitlab.com/ayaka/internal/adapter/rest"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/jwt"
)

func RegisterDatabase() {
	appContainer.RegisterService("database", new(repository.Sqlx))
}

func RegisterCache() {
	appContainer.RegisterService("cache", new(repository.Cache))
}

func RegisterRest() {
	appContainer.RegisterService("fiber", new(rest.Fiber))
}

func RegisterRepository() {
	appContainer.RegisterService("tblUserRepository", new(sqlx.TblUserRepository))
	appContainer.RegisterService("tblUserCacheRepository", new(cache.TblUserRepository))

	appContainer.RegisterService("tblCountryRepository", new(sqlx.TblCountryRepository))

	appContainer.RegisterService("tblProvinceRepository", new(sqlx.TblProvinceRepository))

	appContainer.RegisterService("tblCityRepository", new(sqlx.TblCityRepository))

	appContainer.RegisterService("tblUomRepository", new(sqlx.TblUomRepository))

	appContainer.RegisterService("tblCoaRepository", new(sqlx.TblCoaRepository))

	appContainer.RegisterService("tblItemCatRepository", new(sqlx.TblItemCatRepository))

	appContainer.RegisterService("tblWarehouseRepository", new(sqlx.TblWarehouseRepository))

	appContainer.RegisterService("tblWarehouseCategoryRepository", new(sqlx.TblWarehouseCategoryRepository))

	appContainer.RegisterService("tblLogRepository", new(sqlx.TblLogRepository))

	appContainer.RegisterService("tblItemRepository", new(sqlx.TblItemRepository))

	appContainer.RegisterService("tblCurrencyRepository", new(sqlx.TblCurrencyRepository))

	appContainer.RegisterService("tblInitStockRepository", new(sqlx.TblInitStockRepository))

	appContainer.RegisterService("tblStockAdjustRepository", new(sqlx.TblStockAdjustRepository))

	appContainer.RegisterService("tblStockSummaryRepository", new(sqlx.TblStockSummaryRepository))

	appContainer.RegisterService("tblStockMutationRepository", new(sqlx.TblStockMutationRepository))

	appContainer.RegisterService("tblStockMovementRepository", new(sqlx.TblStockMovementRepository))

	appContainer.RegisterService("tblDirectPurchaseReceiveRepository", new(sqlx.TblDirectPurchaseReceiveRepository))
}

func RegisterHandler() {
	appContainer.RegisterService("jwtHandler", new(jwt.JwtHandler))
	appContainer.RegisterService("middlewareHandler", new(custommiddleware.MiddlewareHandler))
	appContainer.RegisterService("logActivity", new(custommiddleware.LogActivityHandler))
	appContainer.RegisterService("generateID", new(formatid.GenerateIDHandler))
}
