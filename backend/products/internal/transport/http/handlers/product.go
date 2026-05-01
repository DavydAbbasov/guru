package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"guru/backend/products/internal/service"
	customerrors "guru/utils/custom-errors"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Handler interface {
	Register(router fiber.Router)
}

type ProductHandler struct {
	svc      *service.ProductService
	validate *validator.Validate
}

func NewProductHandler(svc *service.ProductService, validate *validator.Validate) *ProductHandler {
	return &ProductHandler{
		svc:      svc,
		validate: validate,
	}
}

func (h *ProductHandler) Register(router fiber.Router) {
	g := router.Group("/products")
	g.Post("/", h.Create)
	g.Delete("/:id", h.Delete)
	g.Get("/", h.List)
}

// Create
//
//	@Summary		Create a product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateProductRequest	true	"Product data"
//	@Success		201		{object}	ProductResponse
//	@Failure		400		{object}	customerrors.ErrorResponse
//	@Failure		500		{object}	customerrors.ErrorResponse
//	@Router			/products [post]
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return customerrors.FiberWriteError(c, fmt.Errorf("invalid request body: %w", customerrors.ErrValidation))
	}

	if err := h.validate.Struct(&req); err != nil {
		return customerrors.FiberWriteError(c, fmt.Errorf("%s: %w", err.Error(), customerrors.ErrValidation))
	}

	product, err := h.svc.Create(c.UserContext(), req.Name)
	if err != nil {
		return customerrors.FiberWriteError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(&ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		CreatedAt: product.CreatedAt,
	})
}

// Delete
//
//	@Summary		Delete a product
//	@Tags			products
//	@Produce		json
//	@Param			id	path		string	true	"Product ID"	format(uuid)
//	@Success		204
//	@Failure		400	{object}	customerrors.ErrorResponse
//	@Failure		404	{object}	customerrors.ErrorResponse
//	@Failure		500	{object}	customerrors.ErrorResponse
//	@Router			/products/{id} [delete]
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return customerrors.FiberWriteError(c, fmt.Errorf("invalid product id: %w", customerrors.ErrValidation))
	}

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return customerrors.FiberWriteError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// List
//
//	@Summary		List products with pagination
//	@Tags			products
//	@Produce		json
//	@Param			page	query		int	false	"Page number"		default(1)
//	@Param			limit	query		int	false	"Items per page"	default(20)
//	@Success		200		{object}	ListProductsResponse
//	@Failure		500		{object}	customerrors.ErrorResponse
//	@Router			/products [get]
func (h *ProductHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", DefaultPage)
	limit := c.QueryInt("limit", DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	products, total, err := h.svc.List(c.UserContext(), page, limit)
	if err != nil {
		return customerrors.FiberWriteError(c, err)
	}

	items := make([]*ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, &ProductResponse{
			ID:        p.ID,
			Name:      p.Name,
			CreatedAt: p.CreatedAt,
		})
	}

	return c.JSON(&ListProductsResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
