package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	agentsvc "github.com/davidakpele/property-marketplace/internal/agent"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
	"github.com/davidakpele/property-marketplace/pkg/validation"
)

type Handler struct {
	svc *agentsvc.Service
}

func NewHandler(svc *agentsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	agents := r.Group("/agents")
	{
		agents.POST("", h.Create)
		agents.GET("", h.List)
		agents.GET("/:id", h.GetByID)
		agents.PUT("/:id", h.Update)
		agents.DELETE("/:id", h.Delete)
	}
}

// Create godoc
// @Summary      Create an agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        agent  body      CreateAgentRequest  true  "Agent payload"
// @Success      201    {object}  DataResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      409    {object}  ErrorResponse
// @Failure      422    {object}  ValidationErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /agents [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, httpx.NewBadRequest("invalid request body"))
		return
	}
	if err := validation.Validate(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	a, err := h.svc.Create(c.Request.Context(), agentsvc.CreateAgentInput{
		Name:   req.Name,
		Email:  req.Email,
		Phone:  req.Phone,
		Agency: req.Agency,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Created(c, toResponse(a))
}

// List godoc
// @Summary      List all agents
// @Tags         agents
// @Produce      json
// @Param        page      query     int  false  "Page number"     default(1)
// @Param        per_page  query     int  false  "Items per page"  default(20)
// @Success      200       {object}  DataListResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /agents [get]
func (h *Handler) List(c *gin.Context) {
	params := pagination.FromContext(c)

	agents, total, err := h.svc.List(c.Request.Context(), params)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OKPaginated(c, toResponseList(agents), pagination.NewMeta(params, total))
}

// GetByID godoc
// @Summary      Get an agent by ID
// @Tags         agents
// @Produce      json
// @Param        id   path      string  true  "Agent UUID"
// @Success      200  {object}  DataResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /agents/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	a, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, toResponse(a))
}

// Update godoc
// @Summary      Update an agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        id     path      string              true  "Agent UUID"
// @Param        agent  body      UpdateAgentRequest  true  "Updated agent payload"
// @Success      200    {object}  DataResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      404    {object}  ErrorResponse
// @Failure      409    {object}  ErrorResponse
// @Failure      422    {object}  ValidationErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /agents/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, httpx.NewBadRequest("invalid request body"))
		return
	}
	if err := validation.Validate(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	a, err := h.svc.Update(c.Request.Context(), id, agentsvc.UpdateAgentInput{
		Name:   req.Name,
		Email:  req.Email,
		Phone:  req.Phone,
		Agency: req.Agency,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, toResponse(a))
}

// Delete godoc
// @Summary      Delete an agent
// @Tags         agents
// @Produce      json
// @Param        id   path      string  true  "Agent UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /agents/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.NoContent(c)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, httpx.NewBadRequest("invalid id format")
	}
	return id, nil
}
