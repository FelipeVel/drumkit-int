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
	rg.GET("", c.GetAll)
	rg.POST("", c.Create)
}

// GetAll handles GET /loads.
// Returns 200 with the list of loads, or 500 on service/repository error.
func (c *LoadController) GetAll(ctx *gin.Context) {
	loads, err := c.svc.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, loads)
}

// Create handles POST /loads.
// Binds the JSON body to CreateLoadRequest (binding tags enforce required fields),
// returns 400 on malformed/missing input, 201 on success, 500 on service error.
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
