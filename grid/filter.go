package grid

import (
	"database/sql"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/server"
	"time"
)

type UserGrid struct {
	orm.Model
	ID     int
	GridID string
	UserID int

	Name    string
	GroupBy query.NullString

	Filters []UserGridFilter
	Sorting []UserGridSort
	Fields  []UserGridField

	Default     bool
	RowsPerPage query.NullInt
}

// DefaultCache of the models.
func (b UserGrid) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (b UserGrid) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

type UserGridFilter struct {
	orm.Model
	ID         int
	UserGridID int

	Key   string
	Op    string
	Value query.NullString
}

// DefaultCache of the models.
func (b UserGridFilter) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (b UserGridFilter) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

type UserGridSort struct {
	orm.Model
	ID         int
	UserGridID int

	Key  string
	Pos  query.NullInt // because 0 should be allowed as well. TODO figure out a better solution
	Desc bool
}

// DefaultCache of the models.
func (b UserGridSort) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (b UserGridSort) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

type UserGridField struct {
	orm.Model
	ID         int
	UserGridID int

	Key string
	Pos int
}

// DefaultCache of the models.
func (b UserGridField) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (b UserGridField) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

type FeGridFilter struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type FeGridActive struct {
	ID          int      `json:"id,omitempty"`
	RowsPerPage int      `json:"rowsPerPage,omitempty"`
	Sort        []string `json:"sort,omitempty"`
	Group       string   `json:"group,omitempty"`
}

func filterBase(g *grid) (*UserGrid, interface{}, error) {
	claim := g.Controller().Context().Request.JWTClaim().(*auth.Claim)

	userGrid := &UserGrid{}
	err := userGrid.Init(userGrid)
	if err != nil {
		return nil, 0, err
	}

	return userGrid, claim.UserID(), nil
}

func getFilterByID(id int, g *grid) (*UserGrid, error) {
	userGrid, userID, err := filterBase(g)
	if err != nil {
		return nil, err
	}
	err = userGrid.First(condition.New().SetWhere("id = ? AND user_id = ? AND grid_id = ?", id, userID, g.config.ID))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return userGrid, nil
}

func getFilterList(g *grid) ([]FeGridFilter, error) {
	userGrid, userID, err := filterBase(g)
	if err != nil {
		return nil, err
	}

	var res []UserGrid
	userGrid.SetPermissions(orm.WHITELIST, "ID", "Name")
	err = userGrid.All(&res, condition.New().SetWhere("user_id = ? AND grid_id = ?", userID, g.config.ID))
	if err != nil {
		return nil, err
	}

	var rv []FeGridFilter
	for _, row := range res {
		rv = append(rv, FeGridFilter{ID: row.ID, Name: row.Name})
	}

	return rv, nil
}
