package core

import (
	"fmt"
	"net/http"

	extgin "github.com/karsto/glew/common/ext-gin"
	"github.com/karsto/glew/common/validation"
	"{{.TODOProjectImportPath}}/pkg/api/config"
	"{{.TODOProjectImportPath}}/pkg/api/controllers"
	"{{.TODOProjectImportPath}}/pkg/api/store"

	"github.com/gin-gonic/gin/binding"
)

func GetControllers(cfg *config.Core) ([]extgin.Registerer, error) {
	store := store.NewStore(cfg.DBConfig)

	registered := []extgin.Registerer{
		controllers.NewTenantController(store),
		{{.TODOControllersRegistration}}
	}
	return registered, nil
}

func NewServer(cfg *config.Core) (*http.Server, error) {
	s := extgin.NewStd()

	base := s.Group("/api")
	registered, err := GetControllers(cfg)
	if err != nil {
		return nil, err
	}

	for _, v := range registered {
		v.Register(base)
	}
	binding.Validator = new(validation.DefaultValidator)

	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: s,
	}, nil
}
