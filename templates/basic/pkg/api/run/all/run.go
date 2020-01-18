package all

import (
	"{{.TODOProjectImportPath}}/pkg/api/config"
)

func Run() {
	cfg := config.NewAll(".")
	cfg.Print()

	instance, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}

	if err := instance.ListenAndServe(); err != nil {
		panic(err)
	}
}
