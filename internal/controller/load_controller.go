package controller

import (
	"net/http"

	"github.com/FelipeVel/drumkit-int/internal/dto"
	"github.com/FelipeVel/drumkit-int/internal/service"
	"github.com/gin-gonic/gin"
)

// LoadController handles HTTP requests for the /loads resource.
// It contains no business logic — it binds input, delegates to the service,
// and writes the response.
type LoadController struct {
	svc *service.LoadService
}

// NewLoadController constructs a LoadController with the given service.
func NewLoadController(svc *service.LoadService) *LoadController {
	return &LoadController{svc: svc}
}

// RegisterRoutes attaches the controller's handlers to the provided router group.
func (c *LoadController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/loads", c.GetAll)
	rg.POST("/integrations/webhooks/loads", c.Create)
}

// GetAll handles GET /loads.
//
// @Summary      List all loads
// @Description  Returns all freight loads synced from Turvo.
// @Tags         loads
// @Produce      json
// @Success      200  {array}   dto.LoadResponse
// @Failure      500  {object}  map[string]string
// @Router       /loads [get]
func (c *LoadController) GetAll(ctx *gin.Context) {
	loads, err := c.svc.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, loads)
}

// Create handles POST /integrations/webhooks/loads.
//
// @Summary      Create a load via webhook
// @Description  Receives a freight load payload and creates it in Turvo.
// @Tags         loads
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateLoadRequest  true  "Load payload"
// @Success      201   {object}  dto.CreateLoadResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /integrations/webhooks/loads [post]
func (c *LoadController) Create(ctx *gin.Context) {
	var req dto.CreateLoadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := c.svc.Create(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, created)
}
