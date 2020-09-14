package glew

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/iancoleman/strcase"
)

type Frontend struct{}

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

// GenerateStoreVueFile - generates glue vue front end data store file that enables api calls to be made by the vue app
func (_ *Frontend) GenerateStoreVueFile(destDir, verticalName string, ctx StoreTemplateVueCtx) (FileContainer, error) {
	modelName := strcase.ToLowerCamel(verticalName)
	uiStoreDest := path.Join(destDir, NewPaths().UIStore)
	fileName := fmt.Sprintf("%v.js", modelName)

	b, err := ioutil.ReadFile("templates/ui/store-template.js") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	storeTmpl := string(b)

	content, err := ExecuteTemplate("uiStore", storeTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: uiStoreDest,
		FileName:    fileName,
	}
	return out, nil
}

type StoreTemplateVueCtx struct {
	Resource string
}

func (_ *Frontend) NewStoreTemplateVueCtx(vertical VerticalMeta) (StoreTemplateVueCtx, error) {
	out := StoreTemplateVueCtx{
		Resource: strcase.ToKebab(vertical.Name),
	}
	return out, nil
}

type NewTemplateVueCtx struct {
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
	out := NewTemplateVueCtx{
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
func (_ *Frontend) GenerateNewVueFile(destDir, verticalName string, ctx NewTemplateVueCtx) (FileContainer, error) {
	modelName := strcase.ToLowerCamel(verticalName)
	newVueDest := path.Join(destDir, NewPaths().UIComponents)
	fileName := fmt.Sprintf("new%v.vue", modelName)

	b, err := ioutil.ReadFile("templates/ui/new-template.vue") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	newVuewTmpl := string(b)

	content, err := ExecuteTemplate("uiNewModel", newVuewTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: newVueDest,
		FileName:    fileName,
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

func (frontend *Frontend) NewListTemplateVueCtx(vertical VerticalMeta) (ListTemplateVueCtx, error) {
	fieldsmeta := frontend.GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := ListTemplateVueCtx{
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
func (_ *Frontend) GenerateListVueFile(destDir, verticalName string, ctx ListTemplateVueCtx) (FileContainer, error) {
	modelName := pluralizer.Plural(verticalName)
	modelName = strcase.ToLowerCamel(modelName)
	listVueDest := path.Join(destDir, NewPaths().UIComponents)
	fileName := fmt.Sprintf("%v.vue", modelName)

	b, err := ioutil.ReadFile("templates/ui/list-template.vue") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	listTmpl := string(b)

	content, err := ExecuteTemplate("uiListModel", listTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: listVueDest,
		FileName:    fileName,
	}
	return out, nil
}

type Route struct {
	ResourceName       string
	PluralModelName    string
	TitleCaseModelName string
}

type JSRouterCTX struct {
	Routes []Route
}

func (_ *Frontend) NewJSRouterCTX(verticals []VerticalMeta) (JSRouterCTX, error) {
	out := JSRouterCTX{
		Routes: []Route{},
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
func (_ *Frontend) GenerateJSRouterFile(destDir string, ctx JSRouterCTX) (FileContainer, error) {
	routerDest := path.Join(destDir, NewPaths().UI)
	fileName := fmt.Sprintf("%v.js", "router") // TODO: magic strings

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
		Content:     content,
		Destination: routerDest,
		FileName:    fileName,
	}
	return out, nil
}

type ModelFieldMeta struct {
	FieldRule     string
	FieldName     string
	FieldLabel    string
	FieldType     string
	JSONFieldName string
	ColModifers   string
}
type ListTemplateVueCtx struct {
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
