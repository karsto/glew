package model

import (
	"time"

	"github.com/karsto/glew/common/types"
)

// Tenant - the Tenant struct related to db.tenants
type Tenant struct {
	ID        int            `db:"id" json:"id,omitempty" example:"1"`
	Name      string         `db:"name" json:"name,omitempty"  example:"some tenant"`
	IsActive  bool           `db:"is_active" json:"isActive,omitempty" example:"true"`
	Metadata  types.JSONBMap `db:"metadata" json:"metadata,omitempty" example:""`
	CreatedAt time.Time      `db:"created_at" json:"createdAt,omitempty" example:"2018-06-25T23:37:02.445322Z"`
	UpdatedAt time.Time      `db:"updated_at" json:"updatedAt,omitempty" example:"2018-06-25T23:37:02.445322Z"`
}

// UpdateTenant - the tenant struct used to update tenants
type UpdateTenant struct {
	Name     string         `db:"name" json:"name,omitempty" binding:"min=1,max=100" example:"some tenant"`
	IsActive bool           `db:"is_active" json:"isActive,omitempty" binding:"" example:"true"`
	Metadata types.JSONBMap `db:"metadata" json:"metadata,omitempty" binding:"" example:""`
}
