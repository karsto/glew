package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"{{.TODOProjectImportPath}}/internal/session"
	"{{.TODOProjectImportPath}}/pkg/api/model"
	"{{.TODOProjectImportPath}}/pkg/api/store"
)

type TenantController struct {
	db *store.Store
}

func NewTenantController(db *store.Store) *TenantController {
	c := TenantController{
		db: db,
	}
	return &c
}

func (c *TenantController) Register(service *gin.RouterGroup) {
	router := service.Group("tenant")
	{
		router.GET("/", c.Current)
		router.PUT("/", c.Update)
	}
}

// Current - gets the current tenants data
// @Summary gets the current tenants data
// @Description get the current tenant
// @Tags tenant
// @Accept  json
// @Produce  json
// @Success 200 {object} model.Tenant
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /tenant/ [get]
func (c *TenantController) Current(ctx *gin.Context) {
	s := session.FromContext(ctx)

	m, err := c.db.ReadTenant(s.TenantID)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	if m.ID <= 0 {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, m)
}

// Update - updates a tenant
// @Summary Update a tenant
// @Description Update a tenant
// @Tags tenant
// @Accept  json
// @Produce  json
// @Success 200 {object} model.Tenant
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /tenant/ [put]
func (c *TenantController) Update(ctx *gin.Context) {
	s := session.FromContext(ctx)

	m := model.UpdateTenant{}
	if err := ctx.ShouldBind(&m); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	res, err := c.db.UpdateTenant(s.TenantID, m)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	if res.ID <= 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, m)
}
