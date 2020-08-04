package glew

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/karsto/glew/common/files"
)

type GoType struct {
	Name string
	Type reflect.Type
	Tags reflect.StructTag
}

func (t *GoType) GetKey() string {
	out := t.Name + " " + t.Type.Name() + " " + string(t.Tags)
	return out
}

func (t *GoType) GetNewStatement() string {
	out := t.Type.Name() + "{}"
	return out
}

func (t *GoType) IsNillable() bool {
	switch t.Type.String() {

	case "int", "string", "bool", "Time", "time.Time":
		// TODO: better way to do this
		// TODO: time fields
		return false
	}
	return true
}

func (t *GoType) IsNumeric() bool {
	switch t.Type.String() {

	case "int", "int8", "int16", "int32",
		"int64", "uint", "uint8", "uint16", "uint32",
		"uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	}
	return false
}

func (t *GoType) IsString() bool {
	return t.Type.String() == "string"
}

type SQLStrings struct {
	Insert      string
	Read        string
	List        string
	Put         string
	Delete      string
	CreateTable string
	DropTable   string
}

/// ***** Input

// ModelMeta - simple struct with name and field information required to describe a model
type ModelMeta struct {
	Name   string
	Fields []GoType
}

func GetMeta(m interface{}) (ModelMeta, error) {
	t := reflect.TypeOf(m)
	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}

	out := []GoType{}
	for _, f := range fields {
		out = append(out, GoType{
			Name: f.Name,
			Type: f.Type,
			Tags: f.Tag,
		})
	}
	modelOut := ModelMeta{
		Name:   t.Name(),
		Fields: out,
	}
	return modelOut, nil
}

type Paths struct {
	DB           string
	Migrations   string
	Core         string
	Controllers  string
	Tests        string
	Store        string
	Model        string
	UI           string
	UIStore      string
	UIComponents string
}

var paths = Paths{
	DB:           "db",
	Migrations:   "db/migrations",
	Core:         "db/migrations/core",
	Controllers:  "pkg/api/controllers",
	Tests:        "pkg/api/tests",
	Store:        "pkg/api/store",
	Model:        "pkg/api/model",
	UI:           "ui",
	UIStore:      "ui/store",
	UIComponents: "ui/components",
}

func NewPaths() Paths {
	return paths
}

type FileContainer struct {
	Destination string
	FileName    string
	Content     string
}

func ReadFiles(source, destDir string) ([]FileContainer, error) {
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
			destination := filepath.Join(destDir, parentPath)
			f := FileContainer{
				Destination: destination,
				FileName:    name,
				Content:     string(data),
			}
			out = append(out, f)
			return nil
		})
	if err != nil {
		return out, err
	}
	return out, nil
}

func WriteFiles(fContainers []FileContainer) error {
	for _, f := range fContainers {
		err := files.WriteFile("./"+f.Destination, f.FileName, f.Content)
		if err != nil {
			return err
		}
	}
	return nil
}
