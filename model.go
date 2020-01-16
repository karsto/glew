package glew

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/karsto/glew/internal/files"
	"github.com/otiai10/copy"
)

type SQLStrings struct {
	Insert      string
	Read        string
	List        string
	Put         string
	Delete      string
	CreateTable string
	DropTable   string
}
type UtilityStrings struct {
	InitFunc string
	TrimFunc string
}

type ModelVertical struct {
	Reflect   SimpleReflect
	SQL       SQLStrings
	Utilities UtilityStrings
}

type SimpleReflect struct {
	Name         string
	IDFieldName  string
	Fields       []reflect.StructField
	Tags         map[string]reflect.StructTag
	DBFields     []reflect.StructField
	NilFields    []reflect.StructField
	StringFields []reflect.StructField
}

func GetMeta(m interface{}) (SimpleReflect, error) {
	t := reflect.TypeOf(m)

	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}

	dbFields := []reflect.StructField{}
	nilFields := []reflect.StructField{}
	stringFields := []reflect.StructField{}
	for _, f := range fields {
		if v, found := tagMap[f.Name]; found && v.Get("db") != "" {
			dbFields = append(dbFields, f)
		}
		// //TODO: assume if type is not primitive then its nilable and print if structure
		if f.Type.String() == "string" {
			nilFields = append(nilFields, f)
		}
		if f.Type.String() == "string" {
			stringFields = append(stringFields, f)
		}
	}

	out := SimpleReflect{
		Name:         t.Name(),
		Fields:       fields,
		Tags:         tagMap,
		DBFields:     dbFields,
		NilFields:    nilFields,
		StringFields: stringFields,
	}
	return out, nil
}

type Paths struct {
	DB          string
	Migrations  string
	Core        string
	Controllers string
	Store       string
	Model       string
}

func NewPaths() Paths {
	// TODO: save the bits
	return Paths{
		DB:          "db",
		Migrations:  "db/migrations",
		Core:        "db/migrations/core",
		Controllers: "pkg/api/controllers",
		Store:       "pkg/api/store",
		Model:       "pkg/api/model",
	}
}

type FileContainer struct {
	Destination string
	FileName    string
	Content     string
}

func WriteFiles(fContainers []FileContainer) error {
	for _, f := range fContainers {
		err := files.WriteFile(f.Destination, f.FileName, f.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteMigration(v ModelVertical, destDir string, dbScriptIdxStart int, stream []FileContainer) ([]FileContainer, error) {
	migrationsDest := path.Join(destDir, NewPaths().Migrations)
	name := strcase.ToSnake(v.Reflect.Name)

	fileName := fmt.Sprintf("%v_%s.up.sql", dbScriptIdxStart, name)
	stream = append(stream, FileContainer{
		Content:     v.SQL.CreateTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	fileName = fmt.Sprintf("%v_%s.drop.sql", dbScriptIdxStart, name)
	stream = append(stream, FileContainer{
		Content:     v.SQL.DropTable,
		Destination: migrationsDest,
		FileName:    fileName,
	})

	return stream, nil
}

func WriteController(v ModelVertical, destDir string, templateCache map[string]*template.Template, stream []FileContainer) ([]FileContainer, error) {
	controllerName := strcase.ToSnake(v.Reflect.Name)
	constrollerDest := path.Join(destDir, NewPaths().Controllers)
	fileName := fmt.Sprintf("%v.go", controllerName)

	/** EXPECTED CONTEXT
	{{.modelNameTitleCase}}
	{{.modelNamePlural}}
	{{.modelNamePluralTitleCase}}
	{{.modelNameDocs}} // human friendly for docs
	{{.modelIdFieldName}}
	{{.route}}
	*/

	ctx := map[string]interface{}{}

	ctx["modelNameTitleCase"] = strcase.ToCamel(v.Reflect.Name)
	ctx["modelNamePlural"] = ""                                       // TODO:
	ctx["modelNamePluralTitleCase"] = strcase.ToCamel(v.Reflect.Name) // TODO:
	ctx["modelNameDocs"] = strcase.ToDelimited(v.Reflect.Name, ' ')
	ctx["modelIdFieldName"] = "ID"
	ctx["route"] = strcase.ToKebab(v.Reflect.Name)
	const controllerTmpl = `` // TODO: readfile()

	content, err := ExecuteTemplate("storeFunc", controllerTmpl, ctx, templateCache)
	if err != nil {
		return stream, err
	}
	stream = append(stream, FileContainer{
		Content:     content,
		Destination: constrollerDest,
		FileName:    fileName,
	})

	return stream, nil
}

func WriteStore(v ModelVertical, destDir string, templateCache map[string]*template.Template, stream []FileContainer) ([]FileContainer, error) {
	storeName := strcase.ToSnake(v.Reflect.Name)
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)

	/* EXPECTED CONTEXT
	   {{.tableName}}
	   {{.modelNameTitleCase}}
	   {{.modelPropertiesCreate}} // code; args->[]interface{}{*here*}; attached to model; like "m.property";
	   {{.modelPropertiesUpdate}} // code; args->[]interface{}{*here*}; attached to model; like "m.property";
	   {{.sqlInsert}}
	   {{.sqlList}}
	   {{.sqlRead}}
	   {{.sqlUpdate}}
	   {{.sqlDelete}}
	   {{.trimFunc}}
	   {{.initFunc}}
	*/
	ctx := map[string]interface{}{}

	ctx["tableName"] = strcase.ToSnake(v.Reflect.Name)
	ctx["modelNameTitleCase"] = strcase.ToCamel(v.Reflect.Name)
	ctx["modelPropertiesCreate"] = "" // TODO:
	ctx["modelPropertiesUpdate"] = "" // TODO:
	ctx["SQL"] = v.SQL
	ctx["trimFunc"] = v.Utilities.TrimFunc
	ctx["initFunc"] = v.Utilities.InitFunc

	const storeTmpl = `` // TODO: readfile()

	content, err := ExecuteTemplate("storeFunc", storeTmpl, ctx, templateCache)
	if err != nil {
		return stream, err
	}

	stream = append(stream, FileContainer{
		Content:     content,
		Destination: storeDest,
		FileName:    fileName,
	})
	return stream, nil
}

func WriteModel(v ModelVertical, destDir string, stream []FileContainer) ([]FileContainer, error) {
	modelName := strcase.ToSnake(v.Reflect.Name)
	modelDest := path.Join(destDir, NewPaths().Model)
	fileName := fmt.Sprintf("%v.go", modelName)
	stream = append(stream, FileContainer{
		Content:     "TODO:",
		Destination: modelDest,
		FileName:    fileName,
	})

	return stream, nil
}

type Config struct {
	CopyBase         bool
	WriteStore       bool
	WriteMigrations  bool
	WriteControllers bool
	WriteModels      bool
}

func NewConfig() Config {
	return Config{
		CopyBase:         true,
		WriteStore:       true,
		WriteMigrations:  true,
		WriteControllers: true,
		WriteModels:      true,
	}
}

func GenerateAppFromModels(models []interface{}, destDir string) error {
	templateCache := map[string]*template.Template{}

	verticals := []ModelVertical{}
	for _, m := range models {
		vertical, err := GenerateVertical(m, nil, nil, templateCache)
		if err != nil {
			return err
		}
		verticals = append(verticals, vertical)
	}
	err := GenerateApp(verticals, destDir)
	if err != nil {
		return err
	}
	return nil
}

func GenerateApp(verticals []ModelVertical, destDir string) error {
	// copy base
	err := copy.Copy("./base-project", destDir)
	if err != nil {
		return err
	}
	cfg := NewConfig()
	stream := []FileContainer{}
	for _, v := range verticals {
		if cfg.WriteStore {
			stream, err = WriteStore(v, destDir, stream)
			if err != nil {
				return err
			}
		}

		if cfg.WriteMigrations {
			migrationStartId := 2 // TODO: read from directory automatically
			stream, err = WriteMigration(v, destDir, migrationStartId, stream)
			if err != nil {
				return err
			}
		}
		if cfg.WriteControllers {
			stream, err = WriteController(v, destDir, stream)
			if err != nil {
				return err
			}
		}
		if cfg.WriteModels {
			stream, err = WriteModel(v, destDir, stream)
			if err != nil {
				return err
			}
		}
	}

	spew.Dump(stream)

	err = WriteFiles(stream)
	if err != nil {
		panic(err)
	}
	return nil
}

// createM and putM are optional
func GenerateVertical(model, createM, putM interface{}, templateCache map[string]*template.Template) (ModelVertical, error) {
	if createM == nil {
		createM = model
	}
	if putM == nil && createM != nil {
		putM = createM
	}
	if putM == nil {
		putM = model
	}

	out := ModelVertical{}
	modelMeta, err := GetMeta(model)
	if err != nil {
		return out, err
	}

	createMeta, err := GetMeta(createM)
	if err != nil {
		return out, err
	}
	putMeta, err := GetMeta(putM)
	if err != nil {
		return out, err
	}

	ctx := map[string]interface{}{}
	ctx["dbFields"] = modelMeta.DBFields
	ctx["insertFields"] = createMeta.DBFields // insertFields
	ctx["putFields"] = putMeta.DBFields
	ctx["tableName"] = modelMeta.Name // TODO: snake case lower

	utilities, err := getUtilities(templateCache, ctx)
	SQLStrings, err := getSQL(templateCache, ctx)

	out.Reflect = modelMeta
	out.Utilities = utilities
	out.SQL = SQLStrings
	return out, nil
}

// returns m.somestringprop = strings.TrimSpace(m.somestringProp)
func getUtilities(templateCache map[string]*template.Template, ctx map[string]interface{}) (UtilityStrings, error) {
	out := UtilityStrings{}
	const trimTmpl = `
				func trim{{{.modelName}}}(m {{{.modelName}}}) {{{.modelName}}}{
				{{ range _, $value := .stringFields }}
					m.{{.$value.Name}} = strings.TrimSpace(m.{{.$value.Name}})
				{{ end }}
					return m
				}`

	trimUtil, err := ExecuteTemplate("trim", trimTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.TrimFunc = trimUtil

	const nilTmpl = `
	func initialize{{.modelName}}(m {{.modelName}}) {{.modelName}}{
		{{ range _, $value := .nilFields }}
		if m.{{.$value.Name}} == nil {
			m.{{.$value.Name}} = {{.$value.Type.String()}}{{"{}"}}
		}
		{{ end }}
		return m
		}`
	initFunc, err := ExecuteTemplate("nil", nilTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.InitFunc = initFunc
	return out, nil
}

// TODO: handle different model fields for update/create
func getSQL(templateCache map[string]*template.Template, ctx map[string]interface{}) (SQLStrings, error) {
	out := SQLStrings{}
	const insertTmpl = `
	INSERT INTO {{.tableName}} (
				{{ range _, $value := .insertFields }}
				{{$value.Name},
				{{end}}
				VALUES({{range $idx, _ := .insertFields}}${{$.idx}}{{end}})
				RETURNING
				{{.idFieldName}}
	`
	insertSQL, err := ExecuteTemplate("insertSQL", insertTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Insert = insertSQL

	const listTmpl = `
		SELECT
			{{ range _, $value := .dbFields }}
			{{$value.Name},
			{{end}}
		FROM {{.tableName}}
		`
	listSQL, err := ExecuteTemplate("listSQL", listTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.List = listSQL

	const readTmpl = `
	SELECT
		{{ range _, $value := .dbFields }}
		{{$value.Name},
		{{end}}
	FROM  {{.tableName}} WHERE tenant_id = $1 AND {{.idFieldName}} = $2`
	readSQL, err := ExecuteTemplate("readSQL", readTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Read = readSQL

	const putTmpl = `
	UPDATE {{.tableName}} SET
	{{ range idx, $value := .putFields }}
	{{$value.Name} = ${{$idx+2}},
	{{end}}
	WHERE tenant_id = $1 AND {{.idFieldName}} = $2
	`
	putSQL, err := ExecuteTemplate("putSQL", putTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Put = putSQL

	const deleteTmpl = `
	DELETE FROM {{.tableName}} WHERE tenant_id = ? AND {{.idFieldName}} IN (?)
	`
	deleteSQL, err := ExecuteTemplate("deleteSQL", deleteTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.Delete = deleteSQL

	// TODO: add special handling for idField, update_at, created_at
	// TODO: map go type to db type
	const createTblTmpl = `
	CREATE TABLE {{.tableName}} (
		{{.idFieldName}} serial PRIMARY KEY,
		{{ range idx, $value := .dbFields }}
		{{$value.Name} {{$value.dbType}},
		{{end}}
	);`
	createTblSQL, err := ExecuteTemplate("createTblSQL", deleteTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.CreateTable = createTblSQL

	const dropTblTmpl = `DROP TABLE {{.tableName}};`
	dropTblSQL, err := ExecuteTemplate("dropTblSQL", dropTblTmpl, ctx, templateCache)
	if err != nil {
		return out, err
	}
	out.DropTable = dropTblSQL

	return out, err
}

func ExecuteTemplateFile(name, templateName string, ctx map[string]interface{}, templateCache map[string]*template.Template) (string, error) {
	err := InitIfNotFound(name, templateName, templateCache)
	if err != nil {
		return "", err
	}

	tmpl, ok := templateCache[name]
	if !ok {
		return "", fmt.Errorf("template '%s' missing from cache", name)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func ExecuteTemplate(name, templateBody string, ctx map[string]interface{}, templateCache map[string]*template.Template) (string, error) {
	err := InitIfNotFound(name, templateBody, templateCache)
	if err != nil {
		return "", err
	}

	tmpl, ok := templateCache[name]
	if !ok {
		return "", fmt.Errorf("template '%s' missing from cache", name)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func InitIfNotFound(name, templateBody string, templateCache map[string]*template.Template) error {
	_, found := templateCache[name]
	if !found {
		tmpl, err := template.New(name).Parse(templateBody)
		if err != nil {
			return err
		}
		templateCache[name] = tmpl
	}
	return nil
}
