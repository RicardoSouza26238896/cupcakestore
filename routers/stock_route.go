package routers

import (
	"github.com/RicardoSouza26238896/cupcakestore/controllers"
	"github.com/RicardoSouza26238896/cupcakestore/database"
	"github.com/RicardoSouza26238896/cupcakestore/middlewares"
	"github.com/RicardoSouza26238896/cupcakestore/repositories"
	"github.com/RicardoSouza26238896/cupcakestore/services"
	"github.com/gofiber/fiber/v2"
)

type StockRouter struct {
	stockController controllers.StockController
}

func NewStockRouter() *StockRouter {
	// Initialize repositories
	stockRepository := repositories.NewStockRepository(database.DB)

	// Initialize services with repositories
	stockService := services.NewStockService(stockRepository)

	// Initialize controllers with services
	stockController := controllers.NewStockController(stockService)

	return &StockRouter{
		stockController: stockController,
	}
}

func (r *StockRouter) InstallRouters(app *fiber.App) {
	stock := app.Group("/stock").Use(middlewares.LoginAndStaffRequired())

	stock.Get("/create", r.stockController.RenderCreate)
	stock.Post("/create", r.stockController.Create)
	stock.Get("/", r.stockController.RenderStocks)
	stock.Get("/:id", r.stockController.RenderStock)
}
