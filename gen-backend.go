package glew

import (
	"fmt"
	"strings"
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
