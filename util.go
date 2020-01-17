package glew

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/karsto/glew/internal/files"
)

type GoType struct {
	Name string
	Type reflect.Type
	Tags reflect.StructTag
}

func (t *GoType) GetNewStatement() string {
	out := t.Type.Name() + "{}"
	return out
}

func (t *GoType) IsNillable() bool {
	return false // TODO:
}

func (t *GoType) IsString() bool {
	return false // TODO:
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
	DB          string
	Migrations  string
	Core        string
	Controllers string
	Store       string
	Model       string
}

var paths = Paths{
	DB:          "db",
	Migrations:  "db/migrations",
	Core:        "db/migrations/core",
	Controllers: "pkg/api/controllers",
	Store:       "pkg/api/store",
	Model:       "pkg/api/model",
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
