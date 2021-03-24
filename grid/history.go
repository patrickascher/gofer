// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/server"
	"github.com/patrickascher/gofer/slicer"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// HistoryType constants.
const (
	HistoryCreated HistoryType = iota - 1
	HistoryUpdated
	HistoryDeleted
)

var ErrHistoryType = "history type %d is not implemented"
var ErrConvertType = "history convert type %s is not implemented"

// Level type
type HistoryType int32

// String converts the HistoryType code.
func (t HistoryType) String() (string, error) {
	switch t {
	case HistoryCreated:
		return "CREATED", nil
	case HistoryUpdated:
		return "UPDATED", nil
	case HistoryDeleted:
		return "DELETED", nil
	default:
		return "", fmt.Errorf(ErrHistoryType, t)
	}
}

// History struct
type History struct {
	orm.Model
	ID     int
	GridID string
	UserID string
	SrcID  string
	Type   string
	Value  string
}

func (h History) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

func (h History) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

// NewHistory creates a new history entry with the given data.
// Error will return if the HistoryType is unknown.
func NewHistory(GridID string, UserID interface{}, SrcID interface{}, Type HistoryType, Value string) error {
	h := History{}
	err := h.Init(&h)
	if err != nil {
		return err
	}

	h.GridID = GridID
	userId, err := convertToString(UserID)
	if err != nil {
		return err
	}
	h.UserID = userId

	srcId, err := convertToString(SrcID)
	if err != nil {
		return err
	}
	h.SrcID = srcId
	t, err := Type.String()
	if err != nil {
		return err
	}
	h.Type = t
	h.Value = Value

	return h.Create()
}

func convertToString(v interface{}) (string, error) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		var rv []string
		for i := 0; i < reflect.ValueOf(v).Len(); i++ {
			v, err := convertToString(reflect.ValueOf(v).Index(i).Interface())
			if err != nil {
				return "", err
			}
			rv = append(rv, v)
		}
		return strings.Join(rv, "-"), nil
	case reflect.Int:
		return strconv.Itoa(int(reflect.ValueOf(v).Int())), nil
	case reflect.String:
		return v.(string), nil
	}
	return "", fmt.Errorf(ErrConvertType, reflect.TypeOf(v).Kind())
}

// historiesById is a helper to return all history and user entries by the given grid and primary ids.
func historiesById(g *grid) ([]History, []auth.User, error) {
	// get all primary keys.
	pFields := g.PrimaryFields()
	params, err := g.controller.Context().Request.Params()
	if err != nil || len(params) == 0 || len(pFields) == 0 {
		return nil, nil, fmt.Errorf(ErrFieldPrimary, g.config.ID)
	}

	// init history model.
	history := History{}
	var histories []History
	err = history.Init(&history)
	history.SetPermissions(orm.WHITELIST, "Type", "Value", "UserID", "CreatedAt")
	if err != nil {
		return nil, nil, fmt.Errorf(errWrap, err)
	}

	// create condition for all grid ids.
	gridIDs := []interface{}{"%" + g.config.ID + "%"}
	gridIDWhere := "(grid_id LIKE ?"
	for _, gId := range g.config.History.AdditionalIDs {
		gridIDWhere += " OR grid_id LIKE ?"
		gridIDs = append(gridIDs, "%"+gId+"%")
	}
	gridIDWhere += ")"

	// fetch histories.
	err = history.All(&histories, condition.New().SetWhere(gridIDWhere, gridIDs...).SetWhere("src_id = ?", params[pFields[0].referenceName][0]).SetOrder("-created_at"))
	if err != nil {
		return nil, nil, fmt.Errorf(errWrap, err)
	}

	// TODO security to hide hidden fields recursively.
	// Be aware of custom added messages.

	// fetch all users.
	var userIDs []string
	var users []auth.User
	for _, h := range histories {
		if _, exists := slicer.StringExists(userIDs, h.UserID); !exists {
			userIDs = append(userIDs, h.UserID)
		}
	}
	if len(userIDs) > 0 {
		user := auth.User{}
		err = user.Init(&user)
		if err != nil {
			return nil, nil, fmt.Errorf(errWrap, err)
		}
		user.SetPermissions(orm.WHITELIST, "Name", "Surname")
		err = user.All(&users, condition.New().SetWhere("id IN (?)", userIDs))
		if err != nil {
			return nil, nil, fmt.Errorf(errWrap, err)
		}
	}

	return histories, users, nil
}
