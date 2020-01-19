package all

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	extgin "github.com/karsto/glew/common/ext-gin"
	"github.com/karsto/glew/common/validation"

	coresrv "{{.TODOProjectImportPath}}/pkg/api/run/core"
	"{{.TODOProjectImportPath}}/pkg/api/config"
)

func GetControllers(cfg *config.All) ([]extgin.Registerer, error) {
	controllers := []extgin.Registerer{
		{{.TODOControllersRegistration2}}
	}

	coreControllers, err := coresrv.GetControllers(cfg.Core)
	if err != nil {
		return nil, err
	}
	controllers = append(controllers, coreControllers...)

	return controllers, nil
}

func NewServer(cfg *config.All) (*http.Server, error) {
	s := extgin.NewStd()

	controllers, err := GetControllers(cfg)
	if err != nil {
		return nil, err
	}

	base := s.Group("/api")
	for _, v := range controllers {
		v.Register(base)
	}

	binding.Validator = new(validation.DefaultValidator)

	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: s,
	}, nil
}
