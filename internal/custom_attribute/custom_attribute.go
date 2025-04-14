// Package customAttribute handles the management of custom attributes for contacts and conversations.
package customAttribute

import (
	"database/sql"
	"embed"

	"github.com/abhinavxd/libredesk/internal/custom_attribute/models"
	"github.com/abhinavxd/libredesk/internal/dbutil"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-i18n"
	"github.com/lib/pq"
	"github.com/zerodha/logf"
)

var (
	//go:embed queries.sql
	efs embed.FS
)

// Manager manages custom attributes.
type Manager struct {
	q    queries
	lo   *logf.Logger
	i18n *i18n.I18n
}

// Opts contains options for initializing the Manager.
type Opts struct {
	DB   *sqlx.DB
	Lo   *logf.Logger
	I18n *i18n.I18n
}

// queries contains prepared SQL queries.
type queries struct {
	GetCustomAttribute     *sqlx.Stmt `query:"get-custom-attribute"`
	GetAllCustomAttributes *sqlx.Stmt `query:"get-all-custom-attributes"`
	InsertCustomAttribute  *sqlx.Stmt `query:"insert-custom-attribute"`
	DeleteCustomAttribute  *sqlx.Stmt `query:"delete-custom-attribute"`
	UpdateCustomAttribute  *sqlx.Stmt `query:"update-custom-attribute"`
}

// New creates and returns a new instance of the Manager.
func New(opts Opts) (*Manager, error) {
	var q queries
	if err := dbutil.ScanSQLFile("queries.sql", &q, opts.DB, efs); err != nil {
		return nil, err
	}
	return &Manager{
		q:    q,
		lo:   opts.Lo,
		i18n: opts.I18n,
	}, nil
}

// Get retrieves a custom attribute by ID.
func (m *Manager) Get(id int) (models.CustomAttribute, error) {
	var customAttribute models.CustomAttribute
	if err := m.q.GetCustomAttribute.Get(&customAttribute, id); err != nil {
		if err == sql.ErrNoRows {
			return customAttribute, envelope.NewError(envelope.NotFoundError, m.i18n.Ts("globals.messages.notFound", "name", m.i18n.P("globals.terms.customAttribute")), nil)
		}
		m.lo.Error("error fetching custom attribute", "error", err)
		return customAttribute, envelope.NewError(envelope.GeneralError, m.i18n.Ts("globals.messages.errorFetching", "name", m.i18n.P("globals.terms.customAttribute")), nil)
	}
	return customAttribute, nil
}

// GetAll retrieves all custom attributes.
func (m *Manager) GetAll(appliesTo string) ([]models.CustomAttribute, error) {
	var customAttributes = make([]models.CustomAttribute, 0)
	if err := m.q.GetAllCustomAttributes.Select(&customAttributes, appliesTo); err != nil {
		m.lo.Error("error fetching custom attributes", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.Ts("globals.messages.errorFetching", "name", m.i18n.P("globals.terms.customAttribute")), nil)
	}
	return customAttributes, nil
}

// Create creates a new custom attribute.
func (m *Manager) Create(attr models.CustomAttribute) error {
	if _, err := m.q.InsertCustomAttribute.Exec(attr.AppliesTo, attr.Name, attr.Description, attr.Key, pq.Array(attr.Values), attr.DataType, attr.Regex, attr.RegexHint); err != nil {
		if dbutil.IsUniqueViolationError(err) {
			return envelope.NewError(envelope.InputError, m.i18n.Ts("globals.messages.errorAlreadyExists", "name", m.i18n.P("globals.terms.customAttribute")), nil)
		}
		m.lo.Error("error inserting custom attribute", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.Ts("globals.messages.errorCreating", "name", "{globals.terms.customAttribute}"), nil)
	}
	return nil
}

// Update updates a custom attribute by ID.
func (m *Manager) Update(id int, attr models.CustomAttribute) error {
	if _, err := m.q.UpdateCustomAttribute.Exec(id, attr.AppliesTo, attr.Name, attr.Description, pq.Array(attr.Values), attr.DataType, attr.Regex, attr.RegexHint); err != nil {
		m.lo.Error("error updating custom attribute", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.customAttribute}"), nil)
	}
	return nil
}

// Delete deletes a custom attribute by ID.
func (m *Manager) Delete(id int) error {
	if _, err := m.q.DeleteCustomAttribute.Exec(id); err != nil {
		m.lo.Error("error deleting custom attribute", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.Ts("globals.messages.errorDeleting", "name", "{globals.terms.customAttribute}"), nil)
	}
	return nil
}
