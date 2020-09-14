package glew

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/iancoleman/strcase"
)

type Backend struct{}

// GenerateStruct - generates a golang struct.
func (_ *Backend) GenerateStruct(structName string, fields []SField) (string, error) {
	structTpl := `
	type {{.StructName}} struct {
		{{.FieldsStr}}
	}
	`

	fields2 := []string{}
	for _, f := range fields {
		fields2 = append(fields2, fmt.Sprintf("%v %v `%v`", f.Name, f.Type, f.Tags))
	}

	listF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\t%v\n", cur)
	}
	fieldStr := AggStrList(fields2, listF)
	fieldStr = strings.Trim(fieldStr, "\n")
	ctx := map[string]string{
		"StructName": structName,
		"FieldsStr":  fieldStr,
	}
	content, err := ExecuteTemplate("structTpl", structTpl, ctx)
	if err != nil {
		return "", err
	}

	return content, nil
}

// GeneratePage - creates a typed go "Page" model for paging list api endpoints.
func (backend *Backend) GeneratePage(structName string) (string, error) {
	fields := []SField{
		{
			Name: "Records",
			Type: fmt.Sprintf("[]%v", structName),
			Tags: "json:\"records\"",
		},
		{
			Name: "Page",
			Type: "types.PagingInfo",
			Tags: "json:\"page\"",
		},
	}
	return backend.GenerateStruct(structName+"Page", fields)
}

// GenerateTrim - Generates a starter trim function for a given model.
func (_ *Backend) GenerateTrim(structName string, stringFieldNames []string) (string, error) {
	const trimTmpl = `
	func (m *{{.StructName}}) Trim(){ {{ range  $value := .StringFieldNames }}
		m.{{$value}} = strings.TrimSpace(m.{{$value}}){{ end }}
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

// GenerateMapFunc - generates a mapping function between two models that share the same fields.
func (_ *Backend) GenerateMapFunc(structName, targetName string, fields []string) (string, error) {
	toMapPl := `
	func (m {{.StructName}}) To{{.TargetName}}() {{.TargetName}} {
		out := {{.TargetName}}{{print "{}"}}
		{{.MapStatement}}
		return out
	}
	`
	listf := func(idx int, cur, res string) string {
		out := fmt.Sprintf("m.%v = out.%v\n", cur, cur)
		return out
	}
	mapStmt := AggStrList(fields, listf)
	mapStmt = strings.Trim(mapStmt, "\n")
	ctx := map[string]interface{}{
		"StructName":   structName,
		"TargetName":   targetName,
		"MapStatement": mapStmt,
	}
	initFunc, err := ExecuteTemplate("toMapPl", toMapPl, ctx)
	if err != nil {
		return "", err
	}
	return initFunc, nil

}

// GenerateInit - Generates initializer that sets fields to null explicitly.
func (_ *Backend) GenerateInit(structName string, nilStatements map[string]string) (string, error) {
	const nilTmpl = `
	func (m *{{.StructName}}) Initialize() { {{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
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

// GenerateNew - generates a constructor. // TODO: diff with init?
func (_ *Backend) GenerateNew(structName string, nilStatements map[string]string) (string, error) {
	const newTmpl = `
	func New{{.StructName}}()*{{.StructName}} {
		m := {{.StructName}}{}{{ range $key, $value := .NilStatements }}
		if m.{{$key}} == nil {
			m.{{$key}} = {{$value}}
		}
	{{ end }}
	return &m
	}`
	ctx := map[string]interface{}{
		"StructName":    structName,
		"NilStatements": nilStatements,
	}
	initFunc, err := ExecuteTemplate("newTmpl", newTmpl, ctx)
	if err != nil {
		return "", err
	}
	return initFunc, nil
}

// GetCommon - get intersection of two []GoTypes
func (_ *Backend) GetCommon(left, right []GoType) []string {
	common := map[string]bool{}
	for _, v := range left {
		common[v.Name] = false
	}

	for _, v := range right {
		if _, found := common[v.Name]; found {
			common[v.Name] = true
		} else {
			common[v.Name] = false
		}
	}

	out := []string{}
	for k, v := range common {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func (_ *Backend) GetStringFields(fields []GoType) []string {
	out := []string{}
	for _, v := range fields {
		if v.IsString() {
			out = append(out, v.Name)
		}
	}
	return out
}

func (_ *Backend) GetNilableFields(fields []GoType) map[string]string {
	out := map[string]string{}
	for _, v := range fields {
		if v.IsNillable() {
			out[v.Name] = v.GetNewStatement()
		}
	}
	return out
}

func (backend *Backend) NewModelCtx(v VerticalMeta) (ModelCtx, error) {
	fields := []SField{}
	for _, v := range v.Model.Fields {
		fields = append(fields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	mName := v.Name
	model, err := backend.GenerateStruct(mName, fields)
	if err != nil {
		return ModelCtx{}, err
	}

	updateFields := []SField{}
	for _, v := range v.UpdateModel.Fields {
		updateFields = append(updateFields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	updateName := fmt.Sprintf("Update%v", v.Name)
	updateModel, err := backend.GenerateStruct(updateName, updateFields)
	if err != nil {
		return ModelCtx{}, err
	}
	createFields := []SField{}
	for _, v := range v.CreateModel.Fields {
		createFields = append(createFields, SField{
			Name: v.Name,
			Type: v.Type.String(),
			Tags: string(v.Tags),
		})
	}
	createName := fmt.Sprintf("Create%v", v.Name)
	createModel, err := backend.GenerateStruct(createName, createFields)
	if err != nil {
		return ModelCtx{}, err
	}
	utilities := ""
	page, err := backend.GeneratePage(v.Name)
	if err != nil {
		return ModelCtx{}, err
	}
	mTrim, err := backend.GenerateTrim(mName, backend.GetStringFields(v.Model.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	updateTrim, err := backend.GenerateTrim(updateName, backend.GetStringFields(v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	createTrim, err := backend.GenerateTrim(createName, backend.GetStringFields(v.CreateModel.Fields))

	if err != nil {
		return ModelCtx{}, err
	}
	mNil, err := backend.GenerateInit(mName, backend.GetNilableFields(v.Model.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	updateNil, err := backend.GenerateInit(updateName, backend.GetNilableFields(v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	createNil, err := backend.GenerateInit(createName, backend.GetNilableFields(v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	mNew, err := backend.GenerateNew(mName, backend.GetNilableFields(v.Model.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	updateNew, err := backend.GenerateNew(updateName, backend.GetNilableFields(v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	createNew, err := backend.GenerateNew(createName, backend.GetNilableFields(v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	toUpdate, err := backend.GenerateMapFunc(mName, updateName, backend.GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	toCreate, err := backend.GenerateMapFunc(mName, createName, backend.GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	uToModel, err := backend.GenerateMapFunc(updateName, mName, backend.GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	uToCreate, err := backend.GenerateMapFunc(updateName, createName, backend.GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	cToModel, err := backend.GenerateMapFunc(createName, mName, backend.GetCommon(v.Model.Fields, v.CreateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}
	cToUpdate, err := backend.GenerateMapFunc(createName, updateName, backend.GetCommon(v.Model.Fields, v.UpdateModel.Fields))
	if err != nil {
		return ModelCtx{}, err
	}

	utilities = utilities + page + "\n"
	utilities = utilities + toUpdate + "\n"
	utilities = utilities + toCreate + "\n"
	utilities = utilities + uToModel + "\n"
	utilities = utilities + uToCreate + "\n"
	utilities = utilities + cToModel + "\n"
	utilities = utilities + cToUpdate + "\n"
	utilities = utilities + mTrim + "\n"
	utilities = utilities + mNil + "\n"
	utilities = utilities + mNew + "\n"
	utilities = utilities + updateTrim + "\n"
	utilities = utilities + updateNil + "\n"
	utilities = utilities + updateNew + "\n"
	utilities = utilities + createTrim + "\n"
	utilities = utilities + createNil + "\n"
	utilities = utilities + createNew + "\n"

	out := ModelCtx{
		Model:       model,
		CreateModel: createModel,
		UpdateModel: updateModel,
		Utilities:   utilities,
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

func (_ *Backend) NewTestCtx(vertical VerticalMeta) (TestCTX, error) {
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
func (_ *Backend) GenerateRESTTestFile(destDir, verticalName string, ctx TestCTX) (FileContainer, error) {
	modelName := strcase.ToSnake(verticalName)
	testfileDest := path.Join(destDir, NewPaths().Tests)
	fileName := fmt.Sprintf("%v_test.go", modelName)

	b, err := ioutil.ReadFile("templates/crud-api-tests.go") // TODO: no magic strings
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

type ModelCtx struct {
	Model       string
	CreateModel string
	UpdateModel string
	Utilities   string
}

// GenerateModel - generates return, create, update model as well as some boiler plate helper functions - mapping between types, trim, inits
func (_ *Backend) GenerateModel(destDir, verticalName string, ctx ModelCtx) (FileContainer, error) {
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

// Generates all the required
func (_ *Backend) NewStoreCtx(v VerticalMeta, sql SQLStrings, baseCtx BaseAPPCTX) StoreCtx {
	tableName := strcase.ToSnake(v.Name)
	modelNameTitleCase := strcase.ToCamel(v.Name)

	createProperties := []string{}
	for _, v := range v.CreateModel.Fields {
		createProperties = append(createProperties, v.Name)
	}
	listF := func(idx int, cur, res string) string {
		return fmt.Sprintf("\t\tm.%v,\n", cur)
	}
	createProperList := AggStrList(createProperties, listF)
	createProperList = strings.Trim(createProperList, "\n")
	updateProperties := []string{}
	for _, v := range v.UpdateModel.Fields {
		updateProperties = append(updateProperties, v.Name)
	}
	updateProperList := AggStrList(updateProperties, listF)
	updateProperList = strings.Trim(updateProperList, "\n")
	modelNamePluralTitleCase := pluralizer.Plural(modelNameTitleCase)

	out := StoreCtx{
		TODOProjectImportPath:    baseCtx.ImportPath,
		ModelNameTitleCase:       modelNameTitleCase,
		ModelNamePluralTitleCase: modelNamePluralTitleCase,
		TableName:                tableName,
		CreatePropertiesList:     createProperList,
		UpdatePropertiesList:     updateProperList,
		SQL:                      sql,
	}
	return out
}

// GenerateStoreFile - generates a golang ICRUD{{Model}} interface and implementation.
func (_ *Backend) GenerateStoreFile(destDir, verticalName string, ctx StoreCtx) (FileContainer, error) {
	storeName := strcase.ToSnake(verticalName)
	storeDest := path.Join(destDir, NewPaths().Store)
	fileName := fmt.Sprintf("%v.go", storeName)

	b, err := ioutil.ReadFile("templates/crud-store.go")
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

func (_ *Backend) NewControllerCtx(verticalName string, baseCTX BaseAPPCTX) ControllerCtx {
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
func (_ *Backend) GenerateControllerFile(destDir, verticalName string, ctx ControllerCtx) (FileContainer, error) {
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

type SField struct {
	Name string
	Type string
	Tags string
}
