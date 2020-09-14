package glew



type Frontend struct {}

// GenerateFieldMap - generates json field mappings for field:'field'
func  (_ *Frontend) GenerateFieldMap(types []GoType) string {
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
