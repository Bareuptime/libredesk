package models

import (
	"time"

	"github.com/lib/pq"
)

// Default roles
var Roles = []string{
	"Admin",
	"Agent",
}

type Role struct {
	ID          int            `db:"id" json:"id"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Permissions pq.StringArray `db:"permissions" json:"permissions"`
}
