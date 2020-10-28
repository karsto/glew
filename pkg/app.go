package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gertd/go-pluralize"
)

// TODO: simplify names if Create/Update/Result are the same model.

func NewApp(frontend Frontend, backend Backend, db DB) *App {
	return &App{
		db:       db,
		frontend: frontend,
		backend:  backend,
	}
}

type App struct {
	db       DB
	frontend Frontend
	backend  Backend
}

type BaseAPPCTX struct {
	ImportPath string
}

func (_ *App) GetVerticalsFromFile(filePath string) ([]VerticalMeta, error) {

	fset := token.NewFileSet()
	target, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return []VerticalMeta{}, err
	}

	structMap := map[string]VerticalMeta{}
	// get all structs, match / create updates

	ast.Inspect(target, func(n ast.Node) bool {
		// TODO: extract meta info and convert to ModelMeta
		// TODO: consider using ast field types ?

		// TODO: find typespec, then assert typespect.type.(*ast.structtype) and extract relevant info
		// TODO: put ast info in map based on poskey
		// TODO: put comments or check for comments in startPos -1 for meta
		// TODO: figure out how to passthrough everything else and add ignore comments
		/*
					(*ast.TypeSpec)(0xc00019a4e0)({
			 Doc: (*ast.CommentGroup)(<nil>),
			 Name: (*ast.Ident)(0xc00000f920)(TestModel2),
			 Assign: (token.Pos) 0,
			 Type: (*ast.StructType)(0xc00000fba0)({
			  Struct: (token.Pos) 71,
			  Fields: (*ast.FieldList)
		*/
		structBlock, ok := n.(*ast.StructType)
		spec, ok2 := n.(*ast.TypeSpec)

		if ok {
			spew.Dump("structblock", structBlock)
			fmt.Printf("Fields: %v\n", structBlock.Fields)
			fmt.Printf("Struct: %v\n", structBlock.Struct)

		}

		if ok2 {
			spew.Dump("spec", spec)
			fmt.Printf("Comment: %v\n", spec.Comment)
			fmt.Printf("Doc: %v\n", spec.Doc)
			fmt.Printf("Name: %v\n", spec.Name)
			fmt.Printf("Type: %v\n", spec.Type)
		}
		// ast.GenDecl
		// ast.ImportSpec
		// ast.BasicLit
		// ast.GenDecl

		// ast.Comment
		// ast.CommentGroup
		// ast.Node

		// ast.TypeSpec valid
		// ast.FieldList

		// // hard coding looking these up
		// typeDecl := f.Decls[0].(*ast.GenDecl)
		// structDecl := typeDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
		// fields := structDecl.Fields.List

		if ok && ok2 {
			name := spec.Name.Name
			if strings.Contains(name, "Create") {
				name = strings.Replace(name, "Create", "", -1)
				meta, found := structMap[name]
				if !found {
					// make new
				} else {
					fields := []GoType{}
					for _, field := range structBlock.Fields.List {
						fieldSpec := GoType{
							Name: field.Names[0].Name,
							Tags: reflect.StructTag(field.Tag.Value),
							Type: field.Tag.Value,
						}
						fields = append(fields, fieldSpec)
					}
					meta.UpdateName = spec.Name.Name
					meta.UpdateFields = fields
				}
				// TODO: put back in map

			} else if strings.Contains(name, "Update") {
				name = strings.Replace(name, "Update", "", -1)
			} else {

			}

		}
		return true
	})

	result := make([]VerticalMeta, 0, len(structMap))
	for _, v := range structMap {
		if len(v.CreateName) == 0 {
			v.CreateName = v.Name
		}
		if len(v.CreateFields) == 0 {
			v.CreateFields = v.Fields
		}
		if len(v.UpdateName) == 0 {
			v.UpdateName = v.Name
		}
		if len(v.UpdateFields) == 0 {
			v.UpdateFields = v.Fields
		}
		result = append(result, v)
	}

	return result, nil
}

// GenerateVerticalMeta - takes in model, name, and create/update reference models and then using reflection generates the meta for each struct.
// model must be populated. name can be empty, createM, putM can be nil.
func (_ *App) GenerateVerticalMeta(model interface{}, name string, createM, putM interface{}) (VerticalMeta, error) {
	if createM == nil {
		createM = model
	}
	if putM == nil && createM != nil {
		putM = createM
	}
	if putM == nil {
		putM = createM
	}
	modelName, modelFields, err := GetMeta(model)
	if err != nil {
		return VerticalMeta{}, err
	}
	if len(name) < 1 {
		name = modelName
	}

	createName, createFields, err := GetMeta(createM)
	if err != nil {
		return VerticalMeta{}, err
	}
	updateName, updateFields, err := GetMeta(putM)
	if err != nil {
		return VerticalMeta{}, err
	}
	out := VerticalMeta{
		Name:         modelName,
		CreateName:   createName,
		UpdateName:   updateName,
		Fields:       modelFields,
		CreateFields: createFields,
		UpdateFields: updateFields,
	}
	return out, nil
}

// GenerateBaseApp - proccess all the near static templates for the base of the application.
func (_ *App) GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
	files, err := ReadFiles(NewPaths().Static)
	if err != nil {
		return files, err
	}

	basics, err := ReadFiles(NewPaths().BasicTemplates)
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

var pluralizer = pluralize.NewClient() // save the bitsy

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

// GenerateApp - takes in the required information to generate a basic crud app and based on the feature flags enabled, generates those features.
func (app *App) GenerateApp(cfg FeatureConfig, destRoot, appName string, verticals []VerticalMeta, baseCtx BaseAPPCTX) ([]FileContainer, error) {
	// copy base
	// destDir := destRoot //  filepath.Join(destRoot, "base-project")
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

		ctx := app.db.NewSQLCtx(v, migrationStartId)
		sql, err := app.db.GenerateSQL(ctx)
		if err != nil {
			return out, err
		}
		verticalOut.SQL = sql

		if cfg.Store {
			ctx := app.backend.NewStoreCtx(v, sql, baseCtx)
			storeFile, err := app.backend.GenerateStoreFile(ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Store = storeFile
		}

		if cfg.Migrations {
			migrations, err := app.db.GenerateMigrationFiles(sql)
			if err != nil {
				return out, err
			}
			verticalOut.Migrations = append(verticalOut.Migrations, migrations...)
			migrationStartId += 1
		}

		if cfg.Controllers {
			ctx := app.backend.NewControllerCtx(v, baseCtx)
			controllerFile, err := app.backend.GenerateControllerFile(ctx)
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
			model, err := app.backend.GenerateModel(ctx)
			if err != nil {
				return out, err
			}
			verticalOut.Model = model
		}

		if cfg.APICRUDTest {
			ctx, err := app.backend.NewTestCtx(v)
			if err != nil {
				return out, err
			}
			crudTestFile, err := app.backend.GenerateRESTTestFile(ctx)
			if err != nil {
				return out, err
			}
			verticalOut.APICRUDTests = crudTestFile
		}

		if cfg.JSStore {
			ctx, err := app.frontend.NewStoreTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			jsStoreFile, err := app.frontend.GenerateStoreVueFile(ctx)
			if err != nil {
				return out, err
			}
			verticalOut.JSStore = jsStoreFile
		}

		if cfg.VueListModel {
			ctx, err := app.frontend.NewListTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			listFile, err := app.frontend.GenerateListVueFile(ctx)
			if err != nil {
				return out, err
			}
			verticalOut.VueListModel = listFile
		}

		if cfg.VueNewModel {
			ctx, err := app.frontend.NewNewTemplateVueCtx(v)
			if err != nil {
				return out, err
			}
			newVueFile, err := app.frontend.GenerateNewVueFile(ctx)
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
		ctx, err := app.frontend.NewJSRouterCTX(verticals)
		if err != nil {
			return out, err
		}
		jsRouter, err := app.frontend.GenerateJSRouterFile(ctx)
		if err != nil {
			return out, err
		}
		out = append(out, jsRouter)
	}

	return out, nil
}

// Vertical Meta - all the meta information needed to create a vertical.
type VerticalMeta struct {
	Name         string
	UpdateName   string
	CreateName   string
	Fields       []GoType
	CreateFields []GoType
	UpdateFields []GoType
}

// GeneratedVertical - the resulting vertical feature set. Contains Raw strings and objects generated in case they are to be used elsewhere as well as file containers aka digital abstractions of files.
type GeneratedVertical struct {
	SQL        SQLContainer
	Controller FileContainer
	Store      FileContainer
	Model      FileContainer
	Migrations []FileContainer

	JSStore      FileContainer
	VueNewModel  FileContainer
	VueListModel FileContainer
	APICRUDTests FileContainer
}
