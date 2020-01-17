package glew

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

func GenerateBaseApp(destDir, appName string) ([]FileContainer, error) {
	files, err := ReadFiles("./base-project", destDir)
	if err != nil {
		return files, err
	}
	return files, err
}

type ModelCtx struct {
}

func GenerateModel(destDir, verticalName string, ctx ModelCtx) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	modelDest := path.Join(destDir, NewPaths().Model)
	fileName := fmt.Sprintf("%v.go", modelName)
	out := FileContainer{
		Content:     "TODO:",
		Destination: modelDest,
		FileName:    fileName,
	}

	return out, nil
}

type StoreCtx struct {
	TableName                string
	ModelNameTitleCase       string
	ModelNamePluralTitleCase string
	CreateProperties         []string
	UpdateProperties         []string
	SQL                      SQLStrings
}

func GenerateStoreFile(destDir, verticalName string, ctx StoreCtx) (FileContainer, error) {
	storeName := strcase.ToSnake(verticalName)
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)

	b, err := ioutil.ReadFile("example-dal.go.tmpl")
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

func NewControllerCtx(verticalName string) ControllerCtx {
	pluralName := pluralizer.Plural(verticalName)
	out := ControllerCtx{
		ModelNameTitleCase:       strcase.ToCamel(verticalName),
		ModelNamePlural:          pluralName,
		ModelNamePluralTitleCase: strcase.ToCamel(pluralName),
		ModelNameDocs:            strcase.ToDelimited(verticalName, ' '),
		ModelIdFieldName:         "id",
		Route:                    strcase.ToKebab(verticalName),
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
}

// returns Controller name, content, error
func GenerateControllerFile(destDir, verticalName string, ctx ControllerCtx) (FileContainer, error) {
	constrollerDest := path.Join(destDir, NewPaths().Controllers)
	name := strcase.ToSnake(verticalName)
	fileName := fmt.Sprintf("%v.go", name)

	b, err := ioutil.ReadFile("example-controller.go.tmpl")
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
	CopyBase    bool
	Store       bool
	Migrations  bool
	Controllers bool
	Models      bool
	Utilities   bool
}

func NewConfig() FeatureConfig {
	return FeatureConfig{
		CopyBase:    true,
		Store:       true,
		Migrations:  true,
		Controllers: true,
		Models:      true,
	}
}

func GenerateTrim(structName string, stringFieldNames []string) (string, error) {
	const trimTmpl = `
	func trim{{.StructName}}(m model.{{.StructName}}) model.{{.StructName}}}{{{ range  $value := .StringFieldNames }}
		m.{{$value}} = strings.TrimSpace(m.{{$value}}){{ end }}
		return m
	}`
	ctx := map[string]interface{}{
		"StructName":       structName,
		"StringFieldNames": stringFieldNames,
	}
	trimUtil, err := ExecuteTemplate("trim", trimTmpl, ctx)
	if err != nil {
		return "", err
	}
	return trimUtil, nil
}

// nilstatements[fieldname]newStatement
func GenerateInit(structName string, nilStatements map[string]string) (string, error) {
	const nilTmpl = `
	func initialize{{.StructName}}(m model.{{.StructName}}) model.{{.StructName}}{{"{"}}{{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
		return m
	}`
	ctx := map[string]interface{}{
		"StructName":    structName,
		"NilStatements": nilStatements,
	}
	initFunc, err := ExecuteTemplate("niltpl", nilTmpl, ctx)
	if err != nil {
		return "", err
	}
	return initFunc, nil
}

type SQLCtx struct {
	CreateStatements []string
	TableName        string
	InsertFields     []string
	PutFields        []string
	DBFields         []string
	IDColName        string
}

func GenerateSQL(ctx SQLCtx) (SQLStrings, error) {
	out := SQLStrings{}
	const insertTmpl = `
	INSERT INTO {{.TableName}} ({{ range $value := .InsertFields }}
		{{$value}},{{end}}
	VALUES({{range $idx, $val := .InsertFields }},
		${{$idx}},{{end}}
	)
	RETURNING
		{{.IDColName}}
	`
	insertSQL, err := ExecuteTemplate("insertSQL", insertTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.Insert = insertSQL

	const listTmpl = `
		SELECT{{ range $value := .DBFields }}{{$value}},{{end}}
		FROM {{.TableName}}
		`
	listSQL, err := ExecuteTemplate("listSQL", listTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.List = listSQL

	const readTmpl = `
	SELECT{{ range $value := .DBFields }}
		{{$value}},{{end}}
	FROM  {{.TableName}} WHERE tenant_id = $1 AND {{.IDColName}} = $2`
	readSQL, err := ExecuteTemplate("readSQL", readTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.Read = readSQL
	const putTmpl = `
	UPDATE {{.TableName}} SET{{ range $idx, $value := .PutFields }}
		{{$value}} = ${{add $idx 2 }},{{end}}
	WHERE tenant_id = $1 AND {{.IDColName}} = $2
	`
	putSQL, err := ExecuteTemplate("putSQL", putTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.Put = putSQL

	const deleteTmpl = `
	DELETE FROM {{.TableName}} WHERE tenant_id = ? AND {{.IDColName}} IN (?)
	`
	deleteSQL, err := ExecuteTemplate("deleteSQL", deleteTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.Delete = deleteSQL

	const createTblTmpl = `
	CREATE TABLE {{.TableName}} ({{range $idx, $value := .CreateStatements}}{{$length := len .CreateStatements }}
	{{$value}}{{if ne $length $idx }},{{end}}{{end}}
	);`
	createTblSQL, err := ExecuteTemplate("createTblSQL", deleteTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.CreateTable = createTblSQL

	const dropTblTmpl = `DROP TABLE {{.TableName}};`
	dropTblSQL, err := ExecuteTemplate("dropTblSQL", dropTblTmpl, ctx)
	if err != nil {
		return out, err
	}
	out.DropTable = dropTblSQL

	return out, err
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

func GenerateCreateStatement(t DBTypeCtx) string {
	out := strings.Builder{}
	out.WriteString(t.Name)
	out.WriteString(" ")
	out.WriteString(t.Type)
	out.WriteString(" ")
	if !t.IsNullable {
		out.WriteString("NOT NULL ")
	}
	if t.IsPK {
		out.WriteString("PRIMARY KEY ")
	}
	if len(t.Default) > 0 {
		out.WriteString("DEFAULT ")
		out.WriteString(t.Default)
	}
	return out.String()
}
