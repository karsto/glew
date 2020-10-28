package pkg

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/karsto/common/files"
)

type GoType struct {
	Name string
	Type string
	Tags reflect.StructTag
}

func (t *GoType) GetKey() string {
	out := t.Name + " " + t.Type + " " + string(t.Tags)
	return out
}

func (t *GoType) GetNewStatement() string {
	out := t.Type + "{}"
	return out
}

func (t *GoType) IsNillable() bool {
	switch t.Type {

	case "int", "string", "bool", "Time", "time.Time":
		// TODO: better way to do this
		// TODO: time fields
		return false
	}
	return true
}

func (t *GoType) IsNumeric() bool {
	switch t.Type {

	case "int", "int8", "int16", "int32",
		"int64", "uint", "uint8", "uint16", "uint32",
		"uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	}
	return false
}

func (t *GoType) IsString() bool {
	return t.Type == "string"
}

func GetMeta(m interface{}) (string, []GoType, error) {
	t := reflect.TypeOf(m)
	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}

	goTypes := []GoType{}
	for _, f := range fields {
		goTypes = append(goTypes, GoType{
			Name: f.Name,
			Type: f.Type.String(),
			Tags: f.Tag,
		})
	}

	return t.Name(), goTypes, nil
}

type Paths struct {
	DB             string
	Migrations     string
	Core           string
	Controllers    string
	Tests          string
	Store          string
	Model          string
	UI             string
	UIRouter       string
	UIStore        string
	UIComponents   string
	Static         string
	BasicTemplates string
}

var paths = Paths{
	DB:             "db",
	Migrations:     "db/migrations",
	Core:           "db/migrations/core",
	Controllers:    "pkg/api/controllers",
	Tests:          "pkg/api/tests",
	Store:          "pkg/api/store",
	Model:          "pkg/api/model",
	UI:             "ui",
	UIRouter:       "pkg/templates/ui/router-template.js",
	UIStore:        "ui/store",
	UIComponents:   "ui/components",
	Static:         "pkg/static",
	BasicTemplates: "pkg/templates/basic",
}

func NewPaths() Paths {
	return paths
}

type FileContainer struct {
	Path     string
	FileName string
	Content  string
}

func ReadFiles(source string) ([]FileContainer, error) {
	out := []FileContainer{}
	err := filepath.Walk(source,
		// path includes filename
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			name := info.Name()
			nIdx := strings.LastIndex(path, name)
			parentPath := path[:nIdx]
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			parentPath = strings.TrimPrefix(parentPath, source)
			f := FileContainer{
				Path:     parentPath,
				FileName: name,
				Content:  string(data),
			}
			out = append(out, f)
			return nil
		})
	if err != nil {
		return out, err
	}
	return out, nil
}

func WriteFiles(fContainers []FileContainer, dest string) error {
	for _, f := range fContainers {
		path := path.Join("./", dest, f.Path)
		err := files.WriteFile(path, f.FileName, f.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

// AggStrList - runs an aggFunc reduction over strings. func(index, vString, curResultString) -> addition to curResultString
func AggStrList(strs []string, aggFunc func(int, string, string) string) string {
	out := strings.Builder{}
	for i, v := range strs {
		agg := aggFunc(i, v, out.String())
		out.WriteString(agg)
	}
	return out.String()
}
