package core

import (
	"fmt"
	"net/http"

	extgin "github.com/karsto/glew/internal/ext-gin"
	"github.com/karsto/glew/internal/validation"
	"github.com/karsto/base-project/pkg/api/config"
	"github.com/karsto/base-project/pkg/api/controllers"
	"github.com/karsto/base-project/pkg/api/store"

	"github.com/gin-gonic/gin/binding"
)

func GetControllers(cfg *config.Core) ([]extgin.Registerer, error) {
	store := store.NewStore(cfg.DBConfig)

	controllers := []extgin.Registerer{
		controllers.NewTenantController(store),
	}
	return controllers, nil
}

func NewServer(cfg *config.Core) (*http.Server, error) {
	s := extgin.NewStd()

	base := s.Group("/api")
	controllers, err := GetControllers(cfg)
	if err != nil {
		return nil, err
	}

	for _, v := range controllers {
		v.Register(base)
	}
	binding.Validator = new(validation.DefaultValidator)

	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: s,
	}, nil
}
