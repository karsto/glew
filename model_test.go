package glew

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

type TestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type CreateTestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}
type UpdateTestModel struct {
	StringField string    `db:"string_field" json:"stringField"`
	NumField    int       `db:"num_field" json:"numField"`
	ID          int       `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	TimeField   time.Time `db:"time_field" json:"timeField"`
}

func TestPlayground(t *testing.T) {
	/*
		CONFIG
		destDir - the output directory from this run
		appName - the name of the application and docker runtime
		importPath - application go import directory, aka the directory of the app to reference itself
	*/
	app := App{
		db:       DB{},
		frontend: Frontend{},
		backend:  Backend{},
	}
	destDir := "out"
	appName := "testApp"
	importPath := "github.com/ashtonian/glew/out"

	verticals := []VerticalMeta{}

	vertical, err := app.GenerateVerticalMeta(TestModel{}, "TestVertical", CreateTestModel{}, UpdateTestModel{})
	if err != nil {
		panic(err)
	}

	verticals = append(verticals, vertical)

	ctx := BaseAPPCTX{
		ImportPath: importPath, //TODO: hack to produce current import dir + out
	}

	features := NewConfig()
	files, err := app.GenerateApp(features, destDir, appName, verticals, ctx)

	if err != nil {
		panic(err)
	}
	err = WriteFiles(files, destDir)
	if err != nil {
		panic(err)
	}
}

var (
	todoTokens = []string{"TODO", "FIXME"}
)

func isTodo(s string) (int, bool) {
	for _, indent := range todoTokens {
		if strings.HasPrefix(strings.ToUpper(s), indent) {
			return len(indent), true
		}
	}
	return 0, false
}

func TestParsePlayground(t *testing.T) {
	fset := token.NewFileSet()
	target, err := parser.ParseFile(fset, "./test-target.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// ******* Comment Finder
	//TODO: convert to util function to identify all todo places that need dev intervention in final program
	for _, cg := range target.Comments {
		for _, c := range cg.List {
			for i, l := range strings.Split(c.Text, "\n") {
				var t = strings.TrimSpace(l)
				if strings.HasPrefix(t, "//") || strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "*/") {
					t = strings.TrimSpace(t[2:])
				}

				// To do found
				if _, found := isTodo(t); found {
					fmt.Printf("FOUND TODO on line: %v", fset.Position(c.Slash).Line+i)
				}
			}
		}
	}
	// ******** End Comment Finder

	// ******** Struct finder
	ast.Inspect(target, func(n ast.Node) bool {
		// TODO: extract meta info and convert to ModelMeta
		// TODO: consider using ast field types ?
		t, ok2 := n.(*ast.TypeSpec)
		if ok2 {
			fmt.Printf("Name:%v \n", t.Name.Name)
		}
		structBlock, ok := n.(*ast.StructType)
		if ok {
			for _, field := range structBlock.Fields.List {
				fmt.Printf("Field: %s\n", field.Names[0].Name)
				fmt.Printf("Tag:   %s\n", field.Tag.Value)
			}
		}
		return true
	})
}
