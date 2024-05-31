package controllers

import (
	"errors"
	"log"

	"github.com/bitebait/cupcakestore/models"
	"github.com/bitebait/cupcakestore/services"
	"github.com/bitebait/cupcakestore/utils"
	"github.com/bitebait/cupcakestore/views"
	"github.com/gofiber/fiber/v2"
)

type OrderController interface {
	RenderOrder(ctx *fiber.Ctx) error
	RenderAllOrders(ctx *fiber.Ctx) error
	Checkout(ctx *fiber.Ctx) error
	Payment(ctx *fiber.Ctx) error
	RenderCancel(ctx *fiber.Ctx) error
	Cancel(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
}

type orderController struct {
	orderService       services.OrderService
	storeConfigService services.StoreConfigService
}

func NewOrderController(orderService services.OrderService, storeConfigService services.StoreConfigService) OrderController {
	return &orderController{
		orderService:       orderService,
		storeConfigService: storeConfigService,
	}
}

func (c *orderController) Checkout(ctx *fiber.Ctx) error {
	profileID := getUserID(ctx)
	currentUser := ctx.Locals("Profile").(*models.Profile)

	if !currentUser.IsProfileComplete() {
		return renderErrorMessage(ctx, errors.New("Perfil incompleto. Por favor, complete as informações do perfil."), "obter o carrinho de compras")
	}

	cartID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return renderErrorMessage(ctx, err, "processar o ID do carrinho")
	}

	order, err := c.orderService.FindOrCreate(profileID, cartID)
	if err != nil {
		return renderErrorMessage(ctx, err, "obter o carrinho de compras")
	}

	if !c.isAuthorizedUser(currentUser, order, profileID) || !order.CanProceedToCheckout() {
		return ctx.Redirect("/orders")
	}

	storeConfig, err := c.storeConfigService.GetStoreConfig()
	if err != nil {
		return renderErrorMessage(ctx, err, "carregar as configs da loja")
	}

	data := fiber.Map{
		"Order":       order,
		"StoreConfig": storeConfig,
	}
	return views.Render(ctx, "orders/checkout", data, "", storeLayout)
}

func (c *orderController) Payment(ctx *fiber.Ctx) error {
	profileID := getUserID(ctx)
	cartID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return renderErrorMessage(ctx, err, "processar o checkout do carrinho")
	}

	order, err := c.orderService.FindByCartId(cartID)
	if err != nil {
		return renderErrorMessage(ctx, err, "processar o checkout do carrinho")
	}

	currentUser := ctx.Locals("Profile").(*models.Profile)
	if !c.isAuthorizedUser(currentUser, order, profileID) || !order.CanProceedToPayment() {
		return ctx.Redirect("/orders")
	}

	switch ctx.Method() {
	case fiber.MethodPost:
		return c.handlePaymentPostRequest(ctx, order)
	case fiber.MethodGet:
		return c.handlePaymentGetRequest(ctx, order)
	default:
		return ctx.Redirect("/orders")
	}
}

func (c *orderController) isAuthorizedUser(currentUser *models.Profile, order *models.Order, profileID uint) bool {
	return currentUser.User.IsStaff || order.IsCurrentUserOrder(profileID)
}

func (c *orderController) handlePaymentPostRequest(ctx *fiber.Ctx, order *models.Order) error {
	if err := ctx.BodyParser(order); err != nil {
		log.Println(err)
		return renderErrorMessage(ctx, err, "processar os dados de pagamento")
	}

	storeConfig, err := c.storeConfigService.GetStoreConfig()
	if err != nil {
		return renderErrorMessage(ctx, err, "carregar as configs da loja")
	}

	if !storeConfig.DeliveryIsActive {
		order.IsDelivery = false
	}

	if err := c.orderService.Update(order); err != nil {
		return renderErrorMessage(ctx, err, "atualizar o carrinho para pagamento")
	}

	if err := c.orderService.Payment(order); err != nil {
		return renderErrorMessage(ctx, err, "realizar o pagamento do carrinho")
	}

	if order.PaymentMethod == models.PixPaymentMethod {
		return ctx.Redirect("https://pix.ae" + order.PixURL)
	}

	return ctx.Redirect("/orders")
}

func (c *orderController) handlePaymentGetRequest(ctx *fiber.Ctx, order *models.Order) error {
	if order.CanRedirectToPixPayment() {
		return ctx.Redirect("https://pix.ae" + order.PixURL)
	}
	return ctx.Redirect("/orders")
}

func (c *orderController) RenderOrder(ctx *fiber.Ctx) error {
	orderID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return ctx.Redirect("/orders")
	}

	order, err := c.orderService.FindById(orderID)
	if err != nil {
		return ctx.Redirect("/orders")
	}

	currentUser := ctx.Locals("Profile").(*models.Profile)
	if !c.isAuthorizedUser(currentUser, &order, currentUser.ID) {
		return ctx.Redirect("/orders")
	}

	storeConfig, err := c.storeConfigService.GetStoreConfig()
	if err != nil {
		return renderErrorMessage(ctx, err, "carregar configs da loja")
	}

	data := fiber.Map{
		"Order":       order,
		"StoreConfig": storeConfig,
	}
	return views.Render(ctx, "orders/order", data, "", storeLayout)
}

func (c *orderController) RenderAllOrders(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("Profile").(*models.Profile)
	filter := models.NewOrderFilter(currentUser.ID, ctx.QueryInt("page"), ctx.QueryInt("limit"))

	var orders []models.Order
	if currentUser.User.IsStaff {
		orders = c.orderService.FindAll(filter)
	} else {
		orders = c.orderService.FindAllByUser(filter)
	}

	templateName := "orders/orders"
	layout := storeLayout
	if currentUser.User.IsStaff {
		templateName = "orders/admin"
		layout = baseLayout
	}

	return views.Render(ctx, templateName, fiber.Map{"Orders": orders, "Filter": filter}, "", layout)
}

func (c *orderController) RenderCancel(ctx *fiber.Ctx) error {
	orderID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return ctx.Redirect("/orders")
	}

	order, err := c.orderService.FindById(orderID)
	if err != nil {
		return ctx.Redirect("/orders")
	}

	currentUser := ctx.Locals("Profile").(*models.Profile)
	if c.isAuthorizedUser(currentUser, &order, currentUser.ID) {
		return views.Render(ctx, "orders/cancel", order, "", storeLayout)
	}
	return ctx.Redirect("/orders")
}

func (c *orderController) Cancel(ctx *fiber.Ctx) error {
	orderID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return ctx.Redirect("/orders")
	}

	order, err := c.orderService.FindById(orderID)
	if err != nil {
		return ctx.Redirect("/orders")
	}

	currentUser := ctx.Locals("Profile").(*models.Profile)
	if c.isAuthorizedUser(currentUser, &order, currentUser.ID) {
		err = c.orderService.Cancel(order.ID)
		if err != nil {
			return ctx.Redirect("/orders")
		}
	}
	return ctx.Redirect("/orders")
}

func (c *orderController) Update(ctx *fiber.Ctx) error {
	orderID, err := utils.StringToId(ctx.Params("id"))
	if err != nil {
		return ctx.Redirect("/orders")
	}

	order, err := c.orderService.FindById(orderID)
	if err != nil {
		return ctx.Redirect("/orders")
	}

	if order.Status == models.CancelledStatus {
		return ctx.Redirect("/orders")
	}

	order.Status = models.ShoppingCartStatus(ctx.FormValue("status"))

	currentUser := ctx.Locals("Profile").(*models.Profile)
	if currentUser.User.IsStaff {
		if err = c.orderService.Update(&order); err != nil {
			return ctx.Redirect("/orders")
		}
	}
	return ctx.Redirect("/orders")
}
