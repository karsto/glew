package glew

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// TODO: Refactor and consolidate file generate pattern?

type SField struct {
	Name string
	Type string
	Tags string
}

type StoreTemplateVueCtx struct {
	Resource string
}

func NewStoreTemplateVueCtx(vertical VerticalMeta) (StoreTemplateVueCtx, error) {
	out := StoreTemplateVueCtx{
		Resource: strcase.ToKebab(vertical.Name),
	}
	return out, nil
}

func GenerateStoreVueFile(destDir, verticalName string, ctx StoreTemplateVueCtx) (FileContainer, error) {
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

func NewNewTemplateVueCtx(vertical VerticalMeta) (NewTemplateVueCtx, error) {
	modelMeta := GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := NewTemplateVueCtx{
		ModelFieldsMeta:          modelMeta,
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		ModelTitleCaseName:       strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		CamelCasePluralModelName: strcase.ToLowerCamel(pName),
		TitleCaseModelPluralName: strcase.ToCamel(pName),
		FormMapStatment:          GenerateFieldMap(vertical.Model.Fields),          // TODO: LOOP {{.JSONFieldName}}:'{{.JSONFieldName}}',
		FormDefaultStatement:     GenerateFormDefaultsStatement(vertical.Model.Fields), // TODO:   {{.JSONFieldName}}:{{.JSONDefault}}, // default null|''|undefined|false
	}
	return out, nil
}

func GenerateNewVueFile(destDir, verticalName string, ctx NewTemplateVueCtx) (FileContainer, error) {
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

func GetModelFieldMeta(vertical VerticalMeta) []ModelFieldMeta {
	out := []ModelFieldMeta{}
	for _, v := range vertical.Model.Fields {
		mfm := ModelFieldMeta{
			FieldRule:     GetDefaultRule(v),
			FieldName:     strcase.ToLowerCamel(v.Name),
			FieldLabel:    "TODO:" + v.Name,
			FieldType:     GetFieldType(v),
			ColModifers:   GetColMod(v),
			JSONFieldName: strcase.ToLowerCamel(v.Name),
		}
		out = append(out, mfm)
	}
	return out
}

func NewListTemplateVueCtx(vertical VerticalMeta) (ListTemplateVueCtx, error) {
	fieldsmeta := GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := ListTemplateVueCtx{
		ModelFieldsMeta:          fieldsmeta,
		COLOverrideStatement:     GenerateCOLOverrideStatement(vertical.Model.Fields),
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		FormDefaultStatement:     GenerateFormDefaultsStatement(vertical.Model.Fields),
		SearchStatement:          GenerateSearchStatement(vertical.Model.Fields),
		ModelTitleName:           strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		ModelNamePluralTitleCase: strcase.ToCamel(pName),
		CamelCasePlural:          strcase.ToLowerCamel(pName),
	}
	return out, nil
}

func GenerateListVueFile(destDir, verticalName string, ctx ListTemplateVueCtx) (FileContainer, error) {
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

type TestCTX struct {
	ImportPath               string
	ModelNameTitleCase       string
	ModelNamePluralTitleCase string
	ModelNamePluralCamel     string
	DefaultFieldStatement    string // TODO: {{.FieldGOName}}: {{.TODOStringOrINToRGODefault}},
}

func NewTestCtx(vertical VerticalMeta) (TestCTX, error) {
	pName := pluralizer.Plural(vertical.Name)
	out := TestCTX{
		ImportPath:               "",
		ModelNameTitleCase:       vertical.Name,
		ModelNamePluralTitleCase: pName,
		ModelNamePluralCamel:     strcase.ToCamel(pName),
		DefaultFieldStatement:    "//TODO: manually",
	}
	return out, nil
}

func GenerateRESTTestFile(destDir, verticalName string, ctx TestCTX) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	testfileDest := path.Join(destDir, NewPaths().Tests)
	fileName := fmt.Sprintf("%v_test.go", modelName)

	b, err := ioutil.ReadFile("templates/test-template.go") // TODO: no magic strings
	if err != nil {
		return FileContainer{}, err
	}
	testTmpl := string(b)

	content, err := ExecuteTemplate("restCrudTest", testTmpl, ctx) // TODO: magic strings
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: testfileDest,
		FileName:    fileName,
	}
	return out, nil
}

type BaseAPPCTX struct {
	ImportPath string
}

func GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
	files, err := ReadFiles("static", destDir)
	if err != nil {
		return files, err
	}

	basics, err := ReadFiles("templates/basic", destDir)
	if err != nil {
		return files, err
	}

	basicCTX := map[string]string{
		"AppName":                      appName,
		"TODOProjectImportPath":        ctx.ImportPath,
		"TODODockerRegistry":           "{{.TODODockerRegistry}}",
		"TODODockerRepo":               "{{.TODODockerRepo}}",
		"TODOControllersRegistration":  "{{.TODOControllersRegistration}}",
		"TODOControllersRegistration2": "{{.TODOControllersRegistration2}}",
	}

	for i, v := range basics {
		content, err := ExecuteTemplate(fmt.Sprintf("bsc: %v", i)+destDir+v.FileName, v.Content, basicCTX)
		if err != nil {
			return files, err
		}
		basics[i].Content = content
	}
	files = append(files, basics...)

	return files, err
}

type ModelCtx struct {
	Model       string
	CreateModel string
	UpdateModel string
	Utilities   string
}

func GenerateModel(destDir, verticalName string, ctx ModelCtx) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	modelDest := path.Join(destDir, NewPaths().Model)
	fileName := fmt.Sprintf("%v.go", modelName)

	modelTpl := `package model
import (
	"time"
	"github.com/karsto/glew/common/types"
)

{{.Model}}

{{.CreateModel}}

{{.UpdateModel}}

{{.Utilities}}
`
	content, err := ExecuteTemplate("modelFunc", modelTpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: modelDest,
		FileName:    fileName,
	}

	return out, nil
}

type StoreCtx struct {
	TODOProjectImportPath    string
	TableName                string
	ModelNameTitleCase       string
	ModelNamePluralTitleCase string
	CreatePropertiesList     string
	UpdatePropertiesList     string
	SQL                      SQLStrings
}

func GenerateStoreFile(destDir, verticalName string, ctx StoreCtx) (FileContainer, error) {
	storeName := strcase.ToSnake(verticalName)
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)

	b, err := ioutil.ReadFile("templates/example-dal.go")
	if err != nil {
		return FileContainer{}, err
	}
	storeTmpl := string(b)

	content, err := ExecuteTemplate("storeFunc", storeTmpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: storeDest,
		FileName:    fileName,
	}
	return out, nil
}

var pluralizer = pluralize.NewClient() // save the bitsy

func NewControllerCtx(verticalName string, baseCTX BaseAPPCTX) ControllerCtx {
	pluralName := pluralizer.Plural(verticalName)
	out := ControllerCtx{
		ModelNameTitleCase:       strcase.ToCamel(verticalName),
		ModelNamePlural:          pluralName,
		ModelNamePluralTitleCase: strcase.ToCamel(pluralName),
		ModelNameDocs:            strcase.ToDelimited(verticalName, ' '),
		ModelIdFieldName:         "id",
		Route:                    strcase.ToKebab(verticalName),
		TODOProjectImportPath:    baseCTX.ImportPath,
	}
	return out
}

type ControllerCtx struct {
	ModelNameTitleCase       string
	ModelNamePlural          string
	ModelNamePluralTitleCase string
	ModelNameDocs            string
	ModelIdFieldName         string
	Route                    string
	TODOProjectImportPath    string
}

// returns Controller name, content, error
func GenerateControllerFile(destDir, verticalName string, ctx ControllerCtx) (FileContainer, error) {
	constrollerDest := path.Join(destDir, NewPaths().Controllers)
	name := strcase.ToSnake(verticalName)
	fileName := fmt.Sprintf("%v.go", name)

	b, err := ioutil.ReadFile("templates/controller.go")
	if err != nil {
		return FileContainer{}, err
	}
	controllerTmpl := string(b)
	content, err := ExecuteTemplate("controller", controllerTmpl, ctx)
	if err != nil {
		return FileContainer{}, err
	}

	out := FileContainer{
		Content:     content,
		Destination: constrollerDest,
		FileName:    fileName,
	}

	return out, nil
}

type FeatureConfig struct {
	CopyBase     bool
	Store        bool // dal/store/model.go
	Migrations   bool
	Controllers  bool
	Models       bool
	Utilities    bool
	JSStore      bool
	JSRouter     bool
	VueNewModel  bool
	VueListModel bool
	APICRUDTest  bool
}

func NewConfig() FeatureConfig {
	return FeatureConfig{
		CopyBase:     true,
		Store:        true,
		Migrations:   true,
		Controllers:  true,
		Models:       true,
		Utilities:    true,
		JSStore:      true,
		JSRouter:     true,
		VueNewModel:  true,
		VueListModel: true,
		APICRUDTest:  true,
	}
}

type SQLCtx struct {
	CreateStatements []string
	TableName        string
	InsertFields     []string
	PutFields        []string
	DBFields         []string
	IDColName        string
}

type Route struct {
	ResourceName       string
	PluralModelName    string
	TitleCaseModelName string
}

type JSRouterCTX struct {
	Routes []Route
}

func NewJSRouterCTX(verticals []VerticalMeta) (JSRouterCTX, error) {
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

func GenerateJSRouterFile(destDir string, ctx JSRouterCTX) (FileContainer, error) {
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

func GenerateMigrationFiles(destDir, verticalName string, sql SQLStrings, dbScriptIdxStart int) ([]FileContainer, error) {
	migrationsDest := path.Join(destDir, NewPaths().Core)
	verticalName = strcase.ToSnake(verticalName)

	out := []FileContainer{}
	fileName := fmt.Sprintf("%v_%s.up.sql", dbScriptIdxStart, verticalName)
	out = append(out, FileContainer{
		Content:     sql.CreateTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	fileName = fmt.Sprintf("%v_%s.drop.sql", dbScriptIdxStart, verticalName)
	out = append(out, FileContainer{
		Content:     sql.DropTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	return out, nil
}

type DBTypeCtx struct {
	Name       string
	Type       string
	Default    string
	IsPK       bool
	IsNullable bool
}
