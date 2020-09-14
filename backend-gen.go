package glew

type Backend struct {}

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

