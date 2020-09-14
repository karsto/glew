package glew

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

type App struct {
	db       DB
	frontend Frontend
	backend  Backend
}

type StoreTemplateVueCtx struct {
	Resource string
}

func (_ *App) NewStoreTemplateVueCtx(vertical VerticalMeta) (StoreTemplateVueCtx, error) {
	out := StoreTemplateVueCtx{
		Resource: strcase.ToKebab(vertical.Name),
	}
	return out, nil
}

// GenerateStoreVueFile - generates glue vue front end data store file that enables api calls to be made by the vue app
func (_ *App) GenerateStoreVueFile(destDir, verticalName string, ctx StoreTemplateVueCtx) (FileContainer, error) {
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

func (app *App) NewNewTemplateVueCtx(vertical VerticalMeta) (NewTemplateVueCtx, error) {
	modelMeta := app.GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := NewTemplateVueCtx{
		ModelFieldsMeta:          modelMeta,
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		ModelTitleCaseName:       strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		CamelCasePluralModelName: strcase.ToLowerCamel(pName),
		TitleCaseModelPluralName: strcase.ToCamel(pName),
		FormMapStatment:          app.frontend.GenerateFieldMap(vertical.Model.Fields),              // TODO: LOOP {{.JSONFieldName}}:'{{.JSONFieldName}}',
		FormDefaultStatement:     app.frontend.GenerateFormDefaultsStatement(vertical.Model.Fields), // TODO:   {{.JSONFieldName}}:{{.JSONDefault}}, // default null|''|undefined|false
	}
	return out, nil
}

// GenerateNewVueFile - generates a "New Model" form for the vue app.
func (_ *App) GenerateNewVueFile(destDir, verticalName string, ctx NewTemplateVueCtx) (FileContainer, error) {
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

// TODO: move to frontend ?
func (app *App) GetModelFieldMeta(vertical VerticalMeta) []ModelFieldMeta {
	out := []ModelFieldMeta{}
	for _, v := range vertical.Model.Fields {
		mfm := ModelFieldMeta{
			FieldRule:     app.frontend.GetDefaultRule(v),
			FieldName:     strcase.ToLowerCamel(v.Name),
			FieldLabel:    "TODO:" + v.Name,
			FieldType:     app.frontend.GetFieldType(v),
			ColModifers:   app.frontend.GetColMod(v),
			JSONFieldName: strcase.ToLowerCamel(v.Name),
		}
		out = append(out, mfm)
	}
	return out
}

func (app *App) NewListTemplateVueCtx(vertical VerticalMeta) (ListTemplateVueCtx, error) {
	fieldsmeta := app.GetModelFieldMeta(vertical)
	pName := pluralizer.Plural(vertical.Name)
	out := ListTemplateVueCtx{
		ModelFieldsMeta:          fieldsmeta,
		COLOverrideStatement:     app.frontend.GenerateCOLOverrideStatement(vertical.Model.Fields),
		ResourceRoute:            strcase.ToKebab(vertical.Name),
		FormDefaultStatement:     app.frontend.GenerateFormDefaultsStatement(vertical.Model.Fields),
		SearchStatement:          app.frontend.GenerateSearchStatement(vertical.Model.Fields),
		ModelTitleName:           strcase.ToCamel(vertical.Name),
		TitleCaseModelName:       strcase.ToCamel(vertical.Name),
		CamelCaseModelName:       strcase.ToLowerCamel(vertical.Name),
		ModelNamePluralTitleCase: strcase.ToCamel(pName),
		CamelCasePlural:          strcase.ToLowerCamel(pName),
	}
	return out, nil
}

// GenerateListVueFile - Generates a page-able list view vue file for a given model
func (_ *App) GenerateListVueFile(destDir, verticalName string, ctx ListTemplateVueCtx) (FileContainer, error) {
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

func (_ *App) NewTestCtx(vertical VerticalMeta) (TestCTX, error) {
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

// GenerateRESTTestFile - generates a backend api CRUD test file
func (_ *App) GenerateRESTTestFile(destDir, verticalName string, ctx TestCTX) (FileContainer, error) {
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

// GenerateBaseApp - proccess all the near static templates for the base of the application.
func (_ *App) GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
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

// GenerateModel - generates return, create, update model as well as some boiler plate helper functions - mapping between types, trim, inits
func (_ *App) GenerateModel(destDir, verticalName string, ctx ModelCtx) (FileContainer, error) {
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

// GenerateStoreFile - generates a golang ICRUD{{Model}} interface and implementation.
func (_ *App) GenerateStoreFile(destDir, verticalName string, ctx StoreCtx) (FileContainer, error) {
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

func (_ *App) NewControllerCtx(verticalName string, baseCTX BaseAPPCTX) ControllerCtx {
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

// GenerateControllerFile - generates a models gin web controller
func (_ *App) GenerateControllerFile(destDir, verticalName string, ctx ControllerCtx) (FileContainer, error) {
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

func (_ *App) NewJSRouterCTX(verticals []VerticalMeta) (JSRouterCTX, error) {
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
func (_ *App) GenerateJSRouterFile(destDir string, ctx JSRouterCTX) (FileContainer, error) {
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

// GenerateMigrationFiles - generates sql db up and down scripts for a given model.
func (_ *App) GenerateMigrationFiles(destDir, verticalName string, sql SQLStrings, dbScriptIdxStart int) ([]FileContainer, error) {
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

// GenerateApp - takes in the required information to generate a basic crud app and based on the feature flags enabled, generates those features.
func (app *App) GenerateApp(cfg FeatureConfig, destRoot, appName string, verticals []VerticalMeta, baseCtx BaseAPPCTX) ([]FileContainer, error) {
	// copy base
	destDir := destRoot //  filepath.Join(destRoot, "base-project")
	out := []FileContainer{}
	if cfg.CopyBase {
		files, err := app.GenerateBaseApp(destRoot, appName, baseCtx)
		if err != nil {
			return out, err
		}
		out = append(out, files...)
	}

	verticalsOut := []GeneratedVertical{}
	migrationStartId := 2 // TODO: read from directory automatically

	for _, v := range verticals {
		verticalOut := GeneratedVertical{}

		ctx := app.db.NewSQLCtx(v)
		sql, err := app.db.GenerateSQL(ctx)
		if err != nil {
			return out, err
		}
		verticalOut.SQL = sql

		if cfg.Store {
			ctx := NewStoreCtx(v, sql, baseCtx)
			storeFile, err := app.GenerateStoreFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Store = storeFile
		}

		if cfg.Migrations {
			migrations, err := app.GenerateMigrationFiles(destDir, v.Name, sql, migrationStartId)
			if err != nil {
				return out, err
			}
			verticalOut.Migrations = append(verticalOut.Migrations, migrations...)
			migrationStartId += 1
		}

		if cfg.Controllers {
			ctx := app.NewControllerCtx(v.Name, baseCtx)
			controllerFile, err := app.GenerateControllerFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Controller = controllerFile
		}
		if cfg.Models {
			ctx, err := app.backend.NewModelCtx(v)
			if err != nil {
				return out, err
			}
			model, err := app.GenerateModel(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Model = model
		}

		if cfg.APICRUDTest {
			ctx, err := app.NewTestCtx(v)
			if err != nil {
				return out, err
			}
			crudTestFile, err := app.GenerateRESTTestFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.APICRUDTests = crudTestFile
		}

		if cfg.JSStore {
			ctx, err := app.NewStoreTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			jsStoreFile, err := app.GenerateStoreVueFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.JSStore = jsStoreFile
		}

		if cfg.VueListModel {
			ctx, err := app.NewListTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			listFile, err := app.GenerateListVueFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.VueListModel = listFile
		}

		if cfg.VueNewModel {
			ctx, err := app.NewNewTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			newVueFile, err := app.GenerateNewVueFile(destDir, v.Name, ctx)
			if err != nil {
				return out, err
			}
			verticalOut.VueNewModel = newVueFile
		}

		verticalsOut = append(verticalsOut, verticalOut)
	}

	for _, v := range verticalsOut {
		out = append(out, v.Migrations...)
		out = append(out, v.Controller)
		out = append(out, v.Store)
		out = append(out, v.Model)
		out = append(out, v.APICRUDTests)
		out = append(out, v.JSStore)
		out = append(out, v.VueListModel)
		out = append(out, v.VueNewModel)
	}

	if cfg.JSRouter {
		ctx, err := app.NewJSRouterCTX(verticals)
		if err != nil {
			return out, err
		}
		jsRouter, err := app.GenerateJSRouterFile(destDir, ctx)
		if err != nil {
			return out, err
		}
		out = append(out, jsRouter)
	}

	return out, nil
}
