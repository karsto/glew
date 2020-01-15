package controllers

import (
	"fmt"
	"net/http"

	"github.com/a8m/rql"
	extgin "github.com/karsto/duke/internal/ext-gin"
	"github.com/karsto/duke/internal/session"
	"github.com/karsto/duke/internal/types"

	"github.com/gin-gonic/gin"
)
/* Expected context
{.upperCaseModelName}
{.pluralModelname}
{.modelName}
{.idFieldName}
*/

type {.upperCaseModelName}Controller struct {
	db     *store.Store
	parser *rql.Parser
}

func New{.upperCaseModelName}Controller(db *store.Store) *{.upperCaseModelName}Controller {
	parser := rql.MustNewParser(rql.Config{
		Model:         model.{.upperCaseModelName}{},
		FieldSep:      ".",
		DefaultLimit:  20,
		LimitMaxValue: 10000,
		DefaultSort:   []string{"+{.idFieldName}"},
	})
	c := {.upperCaseModelName}Controller{
		db:     db,
		parser: parser,
	}
	return &c
}

func (c *{.upperCaseModelName}Controller) Register(service *gin.RouterGroup) {
	router := service.Group("{.pluralModelname}")
	{
		router.POST("/", c.Create)
		router.GET("/", c.List)
		router.GET("/:{.idFieldName}", c.Read)
		router.PUT("/:{.idFieldName}", c.Update)
		router.DELETE("/*{.idFieldName}", c.Delete)
	}
}

// Create - creates a {.modelName} based on what is passed in via the body
// @Summary Create a {.modelName}
// @Description create a {.modelName}
// @Tags {.pluralModelname}
// @Accept  json
// @Produce  json
// @Param {.modelName} body model.Create{.upperCaseModelName} true "Create {.upperCaseModelName}"
// @Success 201 {object} model.{.upperCaseModelName}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{.pluralModelname}/ [post]
func (c *{.upperCaseModelName}Controller) Create(ctx *gin.Context) {
	s := session.FromContext(ctx)
	m := model.Create{.upperCaseModelName}{}
	if err := ctx.ShouldBind(&m); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	res, err := c.db.Create{.upperCaseModelName}(s.TenantID, m)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

// List - returns a paged result set of {.pluralModelname}
// @Summary List {.pluralModelname}
// @Description get {.pluralModelname}
// @Tags {.pluralModelname}
// @Accept  json
// @Produce  json
// @Param limit query int false "the amount of records  to return per page" default(20) minimum(1) maximum(1000)
// @Param offset query int false "the amount of records to skip when querying" defaults(0) minimum(0)
// @Param sort query []string false "the fields to sort by. Use a `+` or `-` prefix to specify asc or desc" default("+{.idFieldName}")
// @Param filter {object} int false "the filter you would like to search by"

// @Success 200 {object} model.{.upperCaseModelName}sPage
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{.pluralModelname}/ [get]
func (c *{.upperCaseModelName}Controller) List(ctx *gin.Context) {
	s := session.FromContext(ctx)

	q := types.ListRequest{}
	err := ctx.BindQuery(&q)
	if err != nil {
		ctx.Error(fmt.Errorf("query params invalid: %s", err.Error())).SetType(gin.ErrorTypeBind)
		return
	}

	query := rql.Query{Limit: q.Limit, Offset: q.Offset, Sort: q.Sort, Filter: q.Filter}
	params, err := c.parser.ParseQuery(&query)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	m, total, err := c.db.List{.upperCaseModelName}s(s.TenantID, params.Limit, params.Offset, params.Sort, params.FilterExp, params.FilterArgs)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	pageInfo := types.GetPageInfo(q.Offset, q.Limit, total, q.Sort, q.Filter)

	ctx.JSON(http.StatusOK, model.{.upperCaseModelName}sPage{
		Records: m,
		Page:    pageInfo,
	})
}

// Read - reads a {.modelName} by {.idFieldName}
// @Summary read a {.modelName} by {.idFieldName}
// @Description get {.modelName} by {.idFieldName}
// @Tags {.pluralModelname}
// @Accept  json
// @Produce  json
// @Param id path int true "{.modelName}Id" minimum(1)
// @Success 200 {object} model.{.upperCaseModelName}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{.pluralModelname}/{{.idFieldName}} [get]
func (c *{.upperCaseModelName}Controller) Read(ctx *gin.Context) {
	s := session.FromContext(ctx)

	id, err := extgin.ParamInt(ctx, "{.idFieldName}")
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	m, err := c.db.Read{.upperCaseModelName}(s.TenantID, id)
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

// Update - updates a {.modelName}
// @Summary Update a {.modelName}
// @Description Update a {.modelName}
// @Tags {.pluralModelname}
// @Accept  json
// @Produce  json
// @Param  id path int true "{.modelName}Id" minimum(1)
// @Param  {.modelName} body model.Update{.upperCaseModelName} true "Update {.upperCaseModelName}"
// @Success 200 {object} model.{.upperCaseModelName}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{.pluralModelname}/{{.idFieldName}} [put]
func (c *{.upperCaseModelName}Controller) Update(ctx *gin.Context) {
	s := session.FromContext(ctx)

	m := model.Update{.upperCaseModelName}{}
	if err := ctx.ShouldBind(&m); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	id, err := extgin.ParamInt(ctx, "{.idFieldName}")
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	res, err := c.db.Update{.upperCaseModelName}(s.TenantID, id, m)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	if res.ID <= 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// Delete - delete a {.modelName}
// @Summary delete a {.modelName}
// @Description Delete by {.modelName} ID
// @Tags {.pluralModelname}
// @Accept  json
// @Produce  json
// @Param id path int true "{.modelName}Id" Format(int64) minimum(1)
// @Success 204 nil no response
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{.pluralModelname}/{{.idFieldName}} [delete]
func (c *{.upperCaseModelName}Controller) Delete(ctx *gin.Context) {
	s := session.FromContext(ctx)

	q := types.DeleteQuery{}
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	id, err := extgin.ParamOptionalInt(ctx, "{.idFieldName}", 0)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if id > 0 {
		q.IDs = append(q.IDs, id)
	}

	success, err := c.db.Delete{.upperCaseModelName}(s.TenantID, q.IDs)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	if !success {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.Status(http.StatusNoContent)
}
