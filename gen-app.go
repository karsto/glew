package glew

import (
	"fmt"

	"github.com/gertd/go-pluralize"
)

type App struct {
	db       DB
	frontend Frontend
	backend  Backend
}

type BaseAPPCTX struct {
	ImportPath string
}

// GenerateBaseApp - proccess all the near static templates for the base of the application.
func (_ *App) GenerateBaseApp(destDir, appName string, ctx BaseAPPCTX) ([]FileContainer, error) {
	files, err := ReadFiles("static")
	if err != nil {
		return files, err
	}

	basics, err := ReadFiles("templates/basic")
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
	Name        string //TODO: what is name vs model.name
	Model       ModelMeta
	CreateModel ModelMeta
	UpdateModel ModelMeta
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
		putM = model
	}
	modelMeta, err := GetMeta(model)
	if err != nil {
		return VerticalMeta{}, err
	}
	if len(name) < 1 {
		name = modelMeta.Name
	}

	createMeta, err := GetMeta(createM)
	if err != nil {
		return VerticalMeta{}, err
	}
	updateMeta, err := GetMeta(putM)
	if err != nil {
		return VerticalMeta{}, err
	}
	out := VerticalMeta{
		Name:        name,
		Model:       modelMeta,
		CreateModel: createMeta,
		UpdateModel: updateMeta,
	}
	return out, nil
}
