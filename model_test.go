package glew

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestStoreConnectivity(t *testing.T) {

}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

type TestModel struct {
	StringField string `db:"string_field"`
}

func TestPlayground(t *testing.T) {

	/*
		CONFIG
	*/
	destDir := "./out"

	models := []interface{}{
		TestModel{},
	}

	err := GenerateAppFromModels(models, destDir)
	spew.Dump(err)
	if err != nil {
		panic(err)
	}
}
