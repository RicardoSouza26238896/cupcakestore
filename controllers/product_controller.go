package controllers

import (
	"fmt"
	"github.com/bitebait/cupcakestore/models"
	"github.com/bitebait/cupcakestore/services"
	"github.com/bitebait/cupcakestore/utils"
	"github.com/bitebait/cupcakestore/views"
	"github.com/gofiber/fiber/v2"
	"mime/multipart"
	"strconv"
	"strings"
)

type ProductController interface {
	RenderCreate(ctx *fiber.Ctx) error
	HandlerCreate(ctx *fiber.Ctx) error
	RenderProducts(ctx *fiber.Ctx) error
	RenderProduct(ctx *fiber.Ctx) error
}

type productController struct {
	productService services.ProductService
}

func NewProductController(p services.ProductService) ProductController {
	return &productController{
		productService: p,
	}
}

func (c *productController) RenderCreate(ctx *fiber.Ctx) error {
	return views.Render(ctx, "products/create", nil, "", baseLayout)
}

func (c *productController) HandlerCreate(ctx *fiber.Ctx) error {
	product := &models.Product{}
	if err := ctx.BodyParser(product); err != nil {
		errorMessage := "Dados do produto inválidos: " + err.Error()
		return views.Render(ctx, "products/create", nil, errorMessage, baseLayout)
	}

	imageFile, err := ctx.FormFile("image")
	if err != nil {
		return err
	}

	imageFileName, err := generateRandomImageFileName(imageFile)
	if err != nil {
		return err
	}

	imagePath := fmt.Sprintf("./web/images/%s", imageFileName)
	if err := ctx.SaveFile(imageFile, imagePath); err != nil {
		return err
	}

	product.Image = fmt.Sprintf("/images/%s", imageFileName)

	if err := c.productService.Create(product); err != nil {
		errorMessage := "Falha ao criar produto: " + err.Error()
		return views.Render(ctx, "products/create", nil, errorMessage, baseLayout)
	}

	return ctx.Redirect("/products")
}

func generateRandomImageFileName(imageFile *multipart.FileHeader) (string, error) {
	rand := utils.NewRandomizer()
	randString, err := rand.GenerateRandomString(22)
	if err != nil {
		return "", err
	}
	return randString + "." + strings.Split(imageFile.Filename, ".")[1], nil
}

func (c *productController) RenderProducts(ctx *fiber.Ctx) error {
	query := ctx.Query("q", "")

	pagination := models.NewPagination(ctx.QueryInt("page"), ctx.QueryInt("limit"))
	products := c.productService.FindAll(pagination, query)
	data := fiber.Map{
		"Products":   products,
		"Pagination": pagination,
	}

	return views.Render(ctx, "products/products", data, "", baseLayout)
}

func (c *productController) RenderProduct(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	productID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return ctx.Redirect("/products")
	}

	product, err := c.productService.FindById(uint(productID))
	if err != nil {
		return ctx.Redirect("/products")
	}

	return views.Render(ctx, "products/product", product, "", baseLayout)
}
