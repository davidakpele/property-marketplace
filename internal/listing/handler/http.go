package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	listingsvc "github.com/davidakpele/property-marketplace/internal/listing"
	"github.com/davidakpele/property-marketplace/internal/listing/application"
	listingdomain "github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/search"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
	"github.com/davidakpele/property-marketplace/pkg/validation"
)

type Handler struct {
	svc    *listingsvc.Service
	search *search.Service
}

func NewHandler(svc *listingsvc.Service, searchSvc *search.Service) *Handler {
	return &Handler{svc: svc, search: searchSvc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	listings := r.Group("/listings")
	{
		listings.POST("", h.Create)
		listings.GET("", h.List)
		listings.GET("/search", h.Search)
		listings.GET("/:id", h.GetByID)
		listings.PUT("/:id", h.Update)
		listings.DELETE("/:id", h.Delete)
	}
}

// Create godoc
// @Summary      Create a listing
// @Tags         listings
// @Accept       json
// @Produce      json
// @Param        listing  body      CreateListingRequest  true  "Listing payload"
// @Success      201      {object}  DataResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      422      {object}  ValidationErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /listings [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, httpx.NewBadRequest("invalid request body"))
		return
	}
	if err := validation.Validate(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	listing, err := h.svc.Create.Execute(c.Request.Context(), application.CreateListingInput{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Type:        listingdomain.ListingType(req.Type),
		Bedrooms:    req.Bedrooms,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		AgentID:     req.AgentUUID(),
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Created(c, toResponse(listing))
}

// GetByID godoc
// @Summary      Get a listing by ID
// @Tags         listings
// @Produce      json
// @Param        id   path      string  true  "Listing UUID"
// @Success      200  {object}  DataResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /listings/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	listing, err := h.svc.Get.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, toResponse(listing))
}

// List godoc
// @Summary      List all listings
// @Tags         listings
// @Produce      json
// @Param        page      query     int  false  "Page number"     default(1)
// @Param        per_page  query     int  false  "Items per page"  default(20)
// @Success      200       {object}  DataListResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /listings [get]
func (h *Handler) List(c *gin.Context) {
	params := pagination.FromContext(c)

	listings, total, err := h.svc.List.Execute(c.Request.Context(), params)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OKPaginated(c, toResponseList(listings), pagination.NewMeta(params, total))
}

// Search godoc
// @Summary      Search listings
// @Tags         listings
// @Produce      json
// @Param        type       query     string   false  "Listing type"              Enums(rent, sale, shortlet)
// @Param        min_price  query     number   false  "Minimum price"
// @Param        max_price  query     number   false  "Maximum price"
// @Param        bedrooms   query     integer  false  "Number of bedrooms"
// @Param        lat        query     number   false  "Latitude of search origin"
// @Param        lng        query     number   false  "Longitude of search origin"
// @Param        radius_km  query     number   false  "Search radius in kilometres"
// @Param        page       query     int      false  "Page number"               default(1)
// @Param        per_page   query     int      false  "Items per page"            default(20)
// @Success      200        {object}  DataListResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      422        {object}  ValidationErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /listings/search [get]
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.HandleError(c, httpx.NewBadRequest("invalid query parameters"))
		return
	}
	if err := validation.Validate(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	params := pagination.FromContext(c)
	filters := req.ToFilters()

	listings, total, err := h.search.Search(c.Request.Context(), filters, params)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OKPaginated(c, toResponseList(listings), pagination.NewMeta(params, total))
}

// Update godoc
// @Summary      Update a listing
// @Tags         listings
// @Accept       json
// @Produce      json
// @Param        id       path      string                true  "Listing UUID"
// @Param        listing  body      UpdateListingRequest  true  "Updated listing payload"
// @Success      200      {object}  DataResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      422      {object}  ValidationErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /listings/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	var req UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, httpx.NewBadRequest("invalid request body"))
		return
	}
	if err := validation.Validate(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	listing, err := h.svc.Update.Execute(c.Request.Context(), id, application.UpdateListingInput{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Type:        listingdomain.ListingType(req.Type),
		Bedrooms:    req.Bedrooms,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		AgentID:     req.AgentUUID(),
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, toResponse(listing))
}

// Delete godoc
// @Summary      Delete a listing
// @Tags         listings
// @Produce      json
// @Param        id   path      string  true  "Listing UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /listings/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	if err := h.svc.Delete.Execute(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.NoContent(c)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, httpx.NewBadRequest("invalid id format")
	}
	return id, nil
}
