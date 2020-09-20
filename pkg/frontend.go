package pkg

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/iancoleman/strcase"
)

type Frontend struct{}

type ModelFieldMeta struct {
	FieldRule     string
	FieldName     string
	FieldLabel    string
	FieldType     string
	JSONFieldName string
	ColModifers   string
}

// GenerateFieldMap - generates json field mappings for field:'field'
func (_ *Frontend) GenerateFieldMap(types []GoType) string {
	out := strings.Builder{}
	for _, v := range types {
		name := strcase.ToLowerCamel(v.Name)
		stmt := fmt.Sprintf("%s:'%s',", name, name)
		out.WriteString(stmt)
	}
	return out.String()

}

// GetDefaultRule - generates default ui rule for vee-validate
func (_ *Frontend) GetDefaultRule(t GoType) string {
	if t.IsNumeric() {
		return "min_value:1|numeric"
	}
	if t.IsString() {
		return "alpha_dash"
	}
	return "TODO: rule"
}

// GetFieldType - gets the ui field type for a vue input ex: <b-input type="{{.FieldType}}
func (_ *Frontend) GetFieldType(t GoType) string {
	if t.IsNumeric() {
		return "number"
	}
	if t.IsString() {
		return "text"
	}
	return "TODO: field type"
}

// GetColMod - gets the ui column modifiers for the vue list table.
func (_ *Frontend) GetColMod(t GoType) string {
	if t.IsNumeric() {
		return "\nnumberic\nsortable\n"
	}
	if t.IsString() {
		return "\nsortable\n"
	}
	return "TODO: field type"
}

// GenerateCOLOverrideStatement - creates a two way field binding for the fields to map between backend db casing and front end field casing. This is related to a RQL.NET bug TODO:
func (_ *Frontend) GenerateCOLOverrideStatement(fields []GoType) string {
	out := strings.Builder{}
	for _, v := range fields {
		leftCol := fmt.Sprintf("%s:'%s',\n", strcase.ToLowerCamel(v.Name), strcase.ToSnake(v.Name))
		rightCol := fmt.Sprintf("%s:'%s',\n", strcase.ToSnake(v.Name), strcase.ToLowerCamel(v.Name))
		out.WriteString(leftCol)
		out.WriteString(rightCol)
	}
	return out.String()
}

// GenerateFormDefaultsStatement - creates an front end model with default (null) values for use with vue forms.
func (_ *Frontend) GenerateFormDefaultsStatement(fields []GoType) string {
	out := strings.Builder{}
	for _, v := range fields {
		defstmt := fmt.Sprintf("%s: null,\n", strcase.ToLowerCamel(v.Name))
		out.WriteString(defstmt)
	}
	return out.String()
}

// GenerateSearchStatement - creates a front end if structure with default search options using rql statements.
func (_ *Frontend) GenerateSearchStatement(fields []GoType) string {
	out := strings.Builder{}

	for _, v := range fields {
		name := strcase.ToLowerCamel(v.Name)
		if v.IsNumeric() {
			stmt := fmt.Sprintf("\nif (this.search.%s && this.search.%s > 0) {\nfilter.$or.push({ %s: this.search.%s });\n}", name, name, name, name)
			out.WriteString(stmt)
		}
		if v.IsString() {
			txtTmpl := `
			if (this.search.{{.StringField}} && this.search.{{.StringField}}.length > 0) {
			let {{.StringField}} = this.search.{{.StringField}};
			filter.$or.push(
				{
					{{.StringField}}:{$like:` + "`%{{.StringField}}%`}\n}),\n}\n"

			stmt, err := ExecuteTemplate("searchTmpl", txtTmpl, map[string]interface{}{"StringField": name})
			if err != nil {
				println(err)
			}

			out.WriteString(stmt)
		}

	}
	return out.String()
}

type StoreTemplateVueCtx struct {
	FileName string
	Resource string
}

func (_ *Frontend) NewStoreTemplateVueCtx(vertical VerticalMeta) (StoreTemplateVueCtx, error) {
	modelName := strcase.ToLowerCamel(vertical.Name)
	fileName := fmt.Sprintf("%v.js", modelName)
	out := StoreTemplateVueCtx{
		FileName: fileName,
		Resource: strcase.ToKebab(vertical.Name),
	}
	return out, nil
}

// GenerateStoreVueFile - generates glue vue front end data store file that enables api calls to be made by the vue app
func (_ *Frontend) GenerateStoreVueFile(ctx StoreTemplateVueCtx) (FileContainer, error) {
	content, err := ExecuteTemplateFile("templates/ui/store-template.js", "uiStore", ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:  content,
		Path:     NewPaths().UIStore,
		FileName: ctx.FileName,
	}
	return out, nil
}

type NewTemplateVueCtx struct {
	FileName                 string
	ModelFieldsMeta          []ModelFieldMeta
	ResourceRoute            string
	ModelTitleCaseName       string
	TitleCaseModelName       string
	CamelCaseModelName       string
	CamelCasePluralModelName string
	TitleCaseModelPluralName string
	FormMapStatment          string
	FormDefaultStatement     string
}

func (frontend *Frontend) NewNewTemplateVueCtx(vertical VerticalMeta) (NewTemplateVueCtx, error) {
	modelMeta := frontend.GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	modelName := strcase.ToLowerCamel(vertical.Name)

	fileName := fmt.Sprintf("new%v.vue", modelName)
	out := NewTemplateVueCtx{
		FileName:                 fileName,
		ModelFieldsMeta:          modelMeta,
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		ModelTitleCaseName:       strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		CamelCasePluralModelName: strcase.ToLowerCamel(pName),
		TitleCaseModelPluralName: strcase.ToCamel(pName),
		FormMapStatment:          frontend.GenerateFieldMap(vertical.Model.Fields),              // TODO: LOOP {{.JSONFieldName}}:'{{.JSONFieldName}}',
		FormDefaultStatement:     frontend.GenerateFormDefaultsStatement(vertical.Model.Fields), // TODO:   {{.JSONFieldName}}:{{.JSONDefault}}, // default null|''|undefined|false
	}
	return out, nil
}

// GenerateNewVueFile - generates a "New Model" form for the vue app.
func (_ *Frontend) GenerateNewVueFile(ctx NewTemplateVueCtx) (FileContainer, error) {
	content, err := ExecuteTemplateFile("templates/ui/new-template.vue", "uiNewModel", ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:  content,
		Path:     NewPaths().UIComponents,
		FileName: ctx.FileName,
	}
	return out, nil
}

// TODO: move to frontend ?
func (frontend *Frontend) GetModelFieldMeta(vertical VerticalMeta) []ModelFieldMeta {
	out := []ModelFieldMeta{}
	for _, v := range vertical.Model.Fields {
		mfm := ModelFieldMeta{
			FieldRule:     frontend.GetDefaultRule(v),
			FieldName:     strcase.ToLowerCamel(v.Name),
			FieldLabel:    "TODO:" + v.Name,
			FieldType:     frontend.GetFieldType(v),
			ColModifers:   frontend.GetColMod(v),
			JSONFieldName: strcase.ToLowerCamel(v.Name),
		}
		out = append(out, mfm)
	}
	return out
}

type ListTemplateVueCtx struct {
	FileName                 string
	ModelFieldsMeta          []ModelFieldMeta
	ModelTitleName           string
	TitleCaseModelName       string
	CamelCaseModelName       string
	ModelNamePluralTitleCase string
	CamelCasePlural          string
	COLOverrideStatement     string
	ResourceRoute            string
	FormDefaultStatement     string
	SearchStatement          string
}

func (frontend *Frontend) NewListTemplateVueCtx(vertical VerticalMeta) (ListTemplateVueCtx, error) {
	fieldsmeta := frontend.GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)

	modelName := pluralizer.Plural(vertical.Name)
	modelName = strcase.ToLowerCamel(modelName)
	fileName := fmt.Sprintf("%v.vue", modelName)

	out := ListTemplateVueCtx{
		FileName:                 fileName,
		ModelFieldsMeta:          fieldsmeta,
		COLOverrideStatement:     frontend.GenerateCOLOverrideStatement(vertical.Model.Fields),
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		FormDefaultStatement:     frontend.GenerateFormDefaultsStatement(vertical.Model.Fields),
		SearchStatement:          frontend.GenerateSearchStatement(vertical.Model.Fields),
		ModelTitleName:           strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		ModelNamePluralTitleCase: strcase.ToCamel(pName),
		CamelCasePlural:          strcase.ToLowerCamel(pName),
	}
	return out, nil
}

// GenerateListVueFile - Generates a page-able list view vue file for a given model
func (_ *Frontend) GenerateListVueFile(ctx ListTemplateVueCtx) (FileContainer, error) {
	content, err := ExecuteTemplate("templates/ui/list-template.vue", "uiListModel", ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:  content,
		Path:     NewPaths().UIComponents,
		FileName: ctx.FileName,
	}
	return out, nil
}

type Route struct {
	ResourceName       string
	PluralModelName    string
	TitleCaseModelName string
}

type JSRouterCTX struct {
	FileName string
	Routes   []Route
}

func (_ *Frontend) NewJSRouterCTX(verticals []VerticalMeta) (JSRouterCTX, error) {
	out := JSRouterCTX{
		FileName: fmt.Sprintf("%v.js", "router"),
		Routes:   []Route{},
	}

	for _, v := range verticals {
		pluralName := pluralizer.Plural(v.Name)
		route := Route{
			ResourceName:       strcase.ToKebab(v.Name),
			PluralModelName:    strcase.ToLowerCamel(pluralName),
			TitleCaseModelName: v.Name,
		}
		out.Routes = append(out.Routes, route)
	}
	return out, nil
}

// GenerateJSRouterFile - generates a vue router that supports crud operations
func (_ *Frontend) GenerateJSRouterFile(ctx JSRouterCTX) (FileContainer, error) {
	b, err := ioutil.ReadFile("templates/ui/router-template.js") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	routerTmpl := string(b)

	content, err := ExecuteTemplate("uiRouter", routerTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:  content,
		Path:     NewPaths().UI,
		FileName: ctx.FileName,
	}
	return out, nil
}
