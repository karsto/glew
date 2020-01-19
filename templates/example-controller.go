package controllers

import (
	"fmt"
	"net/http"
	"github.com/a8m/rql"
	extgin "github.com/karsto/glew/internal/ext-gin"
	"github.com/karsto/glew/internal/types"
	"github.com/gin-gonic/gin"
	"{{.TODOProjectImportPath}}/internal/session"
	"{{.TODOProjectImportPath}}/pkg/api/store"
	"{{.TODOProjectImportPath}}/pkg/api/model"
)

type {{.ModelNameTitleCase}}Controller struct {
	db     *store.Store
	parser *rql.Parser
}

func New{{.ModelNameTitleCase}}Controller(db *store.Store) *{{.ModelNameTitleCase}}Controller {
	parser := rql.MustNewParser(rql.Config{
		Model:         model.{{.ModelNameTitleCase}}{},
		FieldSep:      ".",
		DefaultLimit:  20,
		LimitMaxValue: 10000,
		DefaultSort:   []string{"+{{.ModelIdFieldName}}"},
	})
	c := {{.ModelNameTitleCase}}Controller{
		db:     db,
		parser: parser,
	}
	return &c
}

func (c *{{.ModelNameTitleCase}}Controller) Register(service *gin.RouterGroup) {
	router := service.Group("{{.Route}}")
	{
		router.POST("/", c.Create)
		router.GET("/", c.List)
		router.GET("/:{{.ModelIdFieldName}}", c.Read)
		router.PUT("/:{{.ModelIdFieldName}}", c.Update)
		router.DELETE("/*{{.ModelIdFieldName}}", c.Delete)
	}
}

// Create - creates a {{.ModelNameDocs}} based on what is passed in via the body
// @Summary Create a {{.ModelNameDocs}}
// @Description create a {{.ModelNameDocs}}
// @Tags {{.ModelNamePlural}}
// @Accept  json
// @Produce  json
// @Param {{.ModelNameDocs}} body model.Create{{.ModelNameTitleCase}} true "Create {{.ModelNameTitleCase}}"
// @Success 201 {object} model.{{.ModelNameTitleCase}}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{{.Route}}/ [post]
func (c *{{.ModelNameTitleCase}}Controller) Create(ctx *gin.Context) {
	s := session.FromContext(ctx)
	m := model.Create{{.ModelNameTitleCase}}{}
	if err := ctx.ShouldBind(&m); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	res, err := c.db.Create{{.ModelNameTitleCase}}(s.TenantID, m)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

// List - returns a paged result set of {{.ModelNamePlural}}
// @Summary List {{.ModelNamePlural}}
// @Description get {{.ModelNamePlural}}
// @Tags {{.ModelNamePlural}}
// @Accept  json
// @Produce  json
// @Param limit query int false "the amount of records  to return per page" default(20) minimum(1) maximum(1000)
// @Param offset query int false "the amount of records to skip when querying" defaults(0) minimum(0)
// @Param sort query []string false "the fields to sort by. Use a `+` or `-` prefix to specify asc or desc" default("+{{.ModelIdFieldName}}")
// @Param filter {object} int false "the filter you would like to search by"

// @Success 200 {object} model.{{.ModelNameTitleCase}}sPage
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{{.Route}}/ [get]
func (c *{{.ModelNameTitleCase}}Controller) List(ctx *gin.Context) {
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

	m, total, err := c.db.List{{.ModelNamePluralTitleCase}}(s.TenantID, params.Limit, params.Offset, params.Sort, params.FilterExp, params.FilterArgs)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	pageInfo := types.GetPageInfo(q.Offset, q.Limit, total, q.Sort, q.Filter)

	ctx.JSON(http.StatusOK, model.{{.ModelNameTitleCase}}sPage{
		Records: m,
		Page:    pageInfo,
	})
}

// Read - reads a {{.ModelNameDocs}} by {{.ModelIdFieldName}}
// @Summary read a {{.ModelNameDocs}} by {{.ModelIdFieldName}}
// @Description get {{.ModelNameDocs}} by {{.ModelIdFieldName}}
// @Tags {{.ModelNamePlural}}
// @Accept  json
// @Produce  json
// @Param id path int true "{{.ModelNameDocs}}Id" minimum(1)
// @Success 200 {object} model.{{.ModelNameTitleCase}}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{{.Route}}/{{ print "{"  .ModelIdFieldName  "}"}} [get]
func (c *{{.ModelNameTitleCase}}Controller) Read(ctx *gin.Context) {
	s := session.FromContext(ctx)

	id, err := extgin.ParamInt(ctx, "{{.ModelIdFieldName}}")
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	m, err := c.db.Read{{.ModelNameTitleCase}}(s.TenantID, id)
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

// Update - updates a {{.ModelNameDocs}}
// @Summary Update a {{.ModelNameDocs}}
// @Description Update a {{.ModelNameDocs}}
// @Tags {{.ModelNamePlural}}
// @Accept  json
// @Produce  json
// @Param  id path int true "{{.ModelNameDocs}}Id" minimum(1)
// @Param  {{.ModelNameDocs}} body model.Update{{.ModelNameTitleCase}} true "Update {{.ModelNameTitleCase}}"
// @Success 200 {object} model.{{.ModelNameTitleCase}}
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{{.Route}}/{{print "{"  .ModelIdFieldName  "}"}} [put]
func (c *{{.ModelNameTitleCase}}Controller) Update(ctx *gin.Context) {
	s := session.FromContext(ctx)

	m := model.Update{{.ModelNameTitleCase}}{}
	if err := ctx.ShouldBind(&m); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	id, err := extgin.ParamInt(ctx, "{{.ModelIdFieldName}}")
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	res, err := c.db.Update{{.ModelNameTitleCase}}(s.TenantID, id, m)
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

// Delete - delete a {{.ModelNameDocs}}
// @Summary delete a {{.ModelNameDocs}}
// @Description Delete by {{.ModelNameDocs}} ID
// @Tags {{.ModelNamePlural}}
// @Accept  json
// @Produce  json
// @Param id path int true "{{.ModelNameDocs}}Id" Format(int64) minimum(1)
// @Success 204 nil no response
// @Failure 400 {object} types.WebError
// @Failure 404 ""
// @Failure 500 {object} types.WebError
// @Router /{{.Route}}/{{print "{"  .ModelIdFieldName  "}"}} [delete]
func (c *{{.ModelNameTitleCase}}Controller) Delete(ctx *gin.Context) {
	s := session.FromContext(ctx)

	q := types.DeleteQuery{}
	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	id, err := extgin.ParamOptionalInt(ctx, "{{.ModelIdFieldName}}", 0)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if id > 0 {
		q.IDs = append(q.IDs, id)
	}

	success, err := c.db.Delete{{.ModelNameTitleCase}}(s.TenantID, q.IDs)
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
