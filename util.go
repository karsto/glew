package main

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"
)

func initIfNotFound(name, templateBody string, templateCache map[string]*template.Template) error {
	_, found := templateCache[name]
	if !found {
		tmpl, err := template.New(name).Parse(templateBody)
		if err != nil {
			return err
		}
		templateCache[name] = tmpl
	}
	return nil
}

func executeTemplate(name, templateBody string, ctx map[string]interface{}, templateCache map[string]*template.Template) (string, error) {
	err := initIfNotFound(name, templateBody, templateCache)
	if err != nil {
		return "", err
	}

	tmpl, ok := templateCache[name]
	if !ok {
		return "", fmt.Errorf("template '%s' missing from cache", name)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// returns model name, fields, and tag map
func getReflect(m interface{}) (string, []reflect.StructField, map[string]reflect.StructTag) {
	t := reflect.TypeOf(m)

	fields := []reflect.StructField{}
	tagMap := map[string]reflect.StructTag{}
	for i := 0; i <= t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
		tagMap[f.Name] = f.Tag
	}
	return t.Name(), fields, tagMap
}
