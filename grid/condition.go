// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

// internal constants.
const (
	conditionSortKey         = "sort"
	conditionSortSeparator   = ","
	conditionFilterPrefix    = "filter_"
	conditionFilterSeparator = ";"
)

// Error messages.
var (
	ErrFieldPrimary    = "grid: primary is not set for %s"
	ErrFieldPermission = "grid: field %s id not allowed to %s or does not exist"
)

// conditionFirst returns a condition for one row by the given primary param.
// It is used in grid mode details, create, update and delete.
// If a grid condition exists, this condition will be appended.
// The controller params will be checked if there is a value for every primary key.
// The Param must have the field.name as key, for the condition the field.referenceID will be added.
// Error will return if not all primary keys are given.
func (g *grid) conditionFirst() (condition.Condition, error) {

	// if no primary or param exists.
	pFields := g.PrimaryFields()
	params, err := g.controller.Context().Request.Params()
	if err != nil || len(params) == 0 || len(pFields) == 0 {
		return nil, fmt.Errorf(ErrFieldPrimary, g.config.ID)
	}

	// create a new condition.
	// copy grid condition, if exists.
	c := condition.New()
	if g.srcCondition != nil {
		c = g.srcCondition.Copy()
	}

	// checking if all primary fields are set by params.
	for _, f := range pFields {
		v, err := g.controller.Context().Request.Param(f.name)
		if err != nil {
			return nil, fmt.Errorf(ErrFieldPrimary, g.config.ID+":"+f.name)
		}
		c.SetWhere(f.referenceID+" = ?", v[0])
	}

	return c, nil
}

func (g *grid) userFilter(c condition.Condition) error {
	filter, err := g.controller.Context().Request.Param("filter")
	if err == nil {
		id, err := strconv.Atoi(filter[0])
		if err != nil {
			return err
		}

		uFilter, err := getFilterByID(id, g)
		if err != nil {
			return err
		}

		// set limit
		if uFilter.RowsPerPage.Valid {
			p, err := g.controller.Context().Request.Params()
			if err != nil {
				return err
			}
			// only add filter config if not manually changed.
			_, exists := p[paginationLimit]
			if !exists {
				p[paginationLimit] = []string{fmt.Sprint(uFilter.RowsPerPage.Int64)}
			}
		}

		// set active Filter
		g.config.Filter.Active.ID = uFilter.ID
		if uFilter.RowsPerPage.Valid {
			g.config.Filter.Active.RowsPerPage = int(uFilter.RowsPerPage.Int64)
		}

		// Position/Disable fields
		if len(uFilter.Fields) > 0 {
			for i, f := range g.Fields() {
				remove := true
				for _, uf := range uFilter.Fields {
					if uf.Key == f.name {
						g.fields[i].SetPosition(uf.Pos)
						remove = false
						break
					}
				}
				g.fields[i].SetRemove(remove).SetHidden(remove)
			}
		}

		// Add filters
		//TODO Mysql,Oracle have different ways to add/sub dates. create a driver based date function.
		for _, f := range uFilter.Filters {
			// custom filters are getting overwritten! Needed because of sub-queries and so on.
			// TODO frontend info
			if _, op, _ := g.Field(f.Key).Filter(); op == "CUSTOM" || op == "CUSTOMLIKE" {
				f.Op = op
			}

			if gridField := g.Field(f.Key); gridField.error == nil && gridField.filterAble {
				switch f.Op {
				case "TODAY":
					c.SetWhere(gridField.filterField + " >= DATENOW")
					c.SetWhere(gridField.filterField + " <= DATENOW")
				case "YESTERDAY":
					c.SetWhere(gridField.filterField + " >= DATENOW-1")
					c.SetWhere(gridField.filterField + " < DATENOW")
				case "WEEK":
					c.SetWhere(gridField.filterField + " = WEEK")
				case "LWEEK":
					c.SetWhere(gridField.filterField + " = WEEK-1")
				case "MONTH":
					c.SetWhere(gridField.filterField + " = MONTH")
				case "LMONTH":
					c.SetWhere(gridField.filterField + " = MONTH-1")
				case "YEAR":
					c.SetWhere(gridField.filterField + " = YEAR")
				case "LYEAR":
					c.SetWhere(gridField.filterField + " = YEAR-1")
				case "!=", "=", ">=", "<=":
					c.SetWhere(gridField.filterField+" "+f.Op+" ?", escape(f.Value.String))
				case "IN":
					c.SetWhere(gridField.filterField+" IN (?)", strings.Split(escape(f.Value.String), conditionFilterSeparator))
				case "NOTIN":
					c.SetWhere(gridField.name+" NOT IN (?)", strings.Split(escape(f.Value.String), conditionFilterSeparator))
				case "NULL":
					c.SetWhere(gridField.filterField + " IS NULL")
				case "NOTNULL":
					c.SetWhere(gridField.filterField + " IS NOT NULL")
				case "LIKE":
					c.SetWhere(gridField.filterField+" LIKE ?", "%%"+escape(f.Value.String)+"%%")
				case "RLIKE":
					c.SetWhere(gridField.filterField+" LIKE ?", escape(f.Value.String)+"%%")
				case "LLIKE":
					c.SetWhere(gridField.filterField+" LIKE ?", "%%"+escape(f.Value.String))
				case query.CUSTOM, query.CUSTOMLIKE:
					var argsCustom []interface{}
					for i := 0; i < strings.Count(gridField.filterField, "?"); i++ {
						if gridField.filterCondition == query.CUSTOMLIKE {
							argsCustom = append(argsCustom, "%%"+escape(f.Value.String)+"%%")
						} else {
							argsCustom = append(argsCustom, escape(f.Value.String))
						}
					}
					c.SetWhere(gridField.filterField, argsCustom...)
				default:
					return fmt.Errorf(ErrFieldPermission, f.Key, "filter")
				}
			} else {
				return fmt.Errorf(ErrFieldPermission, f.Key, "filter")
			}
		}

		// Add sorts
		var sort string
		// add grouping as first param
		if uFilter.GroupBy.Valid {
			sort += uFilter.GroupBy.String
			g.config.Filter.Active.Group = uFilter.GroupBy.String
			//TODO ASC or DESC
		}
		// add order by
		for _, s := range uFilter.Sorting {
			if sort != "" {
				sort += ", "
			}
			op := "ASC"
			if s.Desc {
				op = "DESC"
			}

			if gridField := g.Field(s.Key); gridField.error == nil && gridField.sortAble {
				_, sorting := gridField.Sort()
				sort += sorting + " " + op
				g.config.Filter.Active.Sort = append(g.config.Filter.Active.Sort, s.Key+" "+op)
			} else {
				return fmt.Errorf(ErrFieldPermission, s.Key, "sort")
			}

		}
		if sort != "" {
			c.SetOrder(sort)
		}
	}
	return nil
}

// conditionAll return a condition for the grid table and export view.
// If a grid condition exists, this condition will be appended.
// Sort and filter_ params are checked. (sort=ID,-Name) (filter_ID=1&filter_Name=John;Doe)
// Error will return if the sort/filter_ field does not exist or has no permission.
func (g *grid) conditionAll() (condition.Condition, error) {

	// create a new condition.
	// if a user condition exist, the value will be copied.
	c := condition.New()
	if g.srcCondition != nil {
		c = g.srcCondition.Copy()
	}

	// user filter
	err := g.userFilter(c)
	if err != nil {
		return nil, err
	}

	// get all controller params.
	params, err := g.controller.Context().Request.Params()
	if err != nil {
		return nil, err
	}

	// check if sort or filter param keys exist.
	for key, param := range params {
		if key == conditionSortKey {
			c.Reset(condition.ORDER)
			err := addSortCondition(g, param[0], c)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(key, conditionFilterPrefix) {
			err := addFilterCondition(g, key[len(conditionFilterPrefix):], param, c)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// addFilterCondition adds a where condition with the given params.
// If there is more than one argument, the condition operator IN will be used.
// Error will return if the field does not exist or the field has no permission for filter.
func addFilterCondition(g *grid, field string, params []string, c condition.Condition) error {

	if gridField := g.Field(field); gridField.error == nil && gridField.filterAble && !g.config.Filter.Disable {

		args := strings.Split(escape(params[0]), conditionFilterSeparator)
		if gridField.filterCondition == query.CUSTOM {
			args = strings.Split(escape(params[0]), ",")
		}
		fmt.Println("aaargs:", args)

		// TODO what is with not... conditions - taking care of?
		if len(args) > 1 && gridField.filterCondition != query.IN && gridField.filterCondition != query.NOTIN && gridField.filterCondition != query.CUSTOM {
			gridField.filterCondition = query.IN
		}

		switch gridField.filterCondition {
		case query.ORACLEDATE: // TODO different driver? create for each driver a callback.
			inputDateFormatISO := "2006-01-02T15:04:05Z"
			inputDateFormat := "2006-01-02"
			outputDateFormatISO := "200601021504"
			outputDateFormat := "20060102"
			var t time.Time
			var err error
			if strings.Index(args[0], ",") == -1 {
				// FROM
				if strings.Index(args[0], "T") != -1 {
					t, err = time.Parse(inputDateFormatISO, args[0])
					if err != nil {
						return err
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDDHH24MI') <= "+gridField.filterField, t.Format(outputDateFormatISO))
				} else {
					t, err = time.Parse(inputDateFormat, args[0])
					if err != nil {
						return err
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDD') <= "+gridField.filterField, t.Format(outputDateFormat))
				}
			} else if strings.HasPrefix(args[0], ",") {
				// TO
				if strings.Index(args[0][1:], "T") != -1 {
					t, err = time.Parse(inputDateFormatISO, args[0][1:])
					if err != nil {
						return err
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDDHH24MI') >= "+gridField.filterField, t.Format(outputDateFormatISO))
				} else {
					t, err = time.Parse(inputDateFormat, args[0][1:])
					if err != nil {
						return err
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDD') >= "+gridField.filterField, t.Format(outputDateFormat))
				}
			} else {
				// FROM TO
				if strings.Index(args[0], "T") != -1 {
					if strings.Index(strings.Split(args[0], ",")[0], "T") != -1 {
						t, err = time.Parse(inputDateFormatISO, strings.Split(args[0], ",")[0])
						if err != nil {
							return err
						}
					} else {
						t, err = time.Parse(inputDateFormat, strings.Split(args[0], ",")[0])
						if err != nil {
							return err
						}
					}
					var t1 time.Time
					if strings.Index(strings.Split(args[0], ",")[1], "T") != -1 {
						t1, err = time.Parse(inputDateFormatISO, strings.Split(args[0], ",")[1])
						if err != nil {
							return err
						}
					} else {
						t1, err = time.Parse(inputDateFormat, strings.Split(args[0], ",")[1])
						if err != nil {
							return err
						}
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDDHH24MI') <= "+gridField.filterField+" AND TO_DATE(?,'YYYYMMDDHH24MI') >= "+gridField.filterField, t.Format(outputDateFormatISO), t1.Format(outputDateFormatISO))
				} else {
					t, err = time.Parse(inputDateFormat, strings.Split(args[0], ",")[0])
					if err != nil {
						return err
					}
					t1, err := time.Parse(inputDateFormat, strings.Split(args[0], ",")[1])
					if err != nil {
						return err
					}
					c.SetWhere("TO_DATE(?,'YYYYMMDD') <= "+gridField.filterField+" AND TO_DATE(?,'YYYYMMDD') >= "+gridField.filterField, t.Format(outputDateFormat), t1.Format(outputDateFormat))
				}
			}
		case query.MYSQLDATE:
			inputDateFormat := "2006-01-02"
			outputDateFormat := "2006-01-02"
			var t time.Time
			var err error
			if strings.Index(args[0], ",") == -1 {
				// FROM
				t, err = time.Parse(inputDateFormat, args[0])
				if err != nil {
					return err
				}
				c.SetWhere("DATE("+gridField.filterField+")>= ?", t.Format(outputDateFormat))
			} else if strings.HasPrefix(args[0], ",") {
				// TO
				t, err = time.Parse(inputDateFormat, args[0][1:])
				if err != nil {
					return err
				}
				c.SetWhere("DATE("+gridField.filterField+")<= ?", t.Format(outputDateFormat))
			} else {
				t, err = time.Parse(inputDateFormat, strings.Split(args[0], ",")[0])
				if err != nil {
					return err
				}
				t1, err := time.Parse(inputDateFormat, strings.Split(args[0], ",")[1])
				if err != nil {
					return err
				}
				c.SetWhere("DATE("+gridField.filterField+")>= ? AND DATE("+gridField.filterField+")<= ?", t.Format(outputDateFormat), t1.Format(outputDateFormat))
			}
		case query.DATE:
			fmt.Println("DATE Filter TODO (different drivers?)")
		case query.LIKE, query.NOTLIKE:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, "%%"+args[0]+"%%")
		case query.NULL, query.NOTNULL:
			c.SetWhere(gridField.filterField + " " + gridField.filterCondition)
		case query.IN, query.NOTIN:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, args)
		case query.RIN, query.RNOTIN:
			c.SetWhere(gridField.filterCondition+" "+gridField.filterField, args)
		case query.SANITIZE:
			if strings.HasPrefix(args[0], "\\%") {
				c.SetWhere("UPPER("+gridField.filterField+") LIKE ?", "%%"+strings.ToUpper(args[0][2:]))
			} else if strings.HasSuffix(args[0], "\\%") {
				c.SetWhere("UPPER("+gridField.filterField+") LIKE ?", strings.ToUpper(args[0][:len(args[0])-2])+"%%")
			} else {
				//TODO check % in text because its escaped by \ from golang.
				c.SetWhere("UPPER("+gridField.filterField+") = ?", strings.ToUpper(args[0]))
			}
		case query.CUSTOM, query.CUSTOMLIKE:
			var argsCustom []interface{}
			for i := 0; i < strings.Count(gridField.filterField, "?"); i++ {
				if gridField.filterCondition == query.CUSTOMLIKE {
					argsCustom = append(argsCustom, "%%"+args[i]+"%%")
				} else {
					argsCustom = append(argsCustom, args[i])
				}
			}
			fmt.Println(gridField.filterField, argsCustom)
			c.SetWhere(gridField.filterField, argsCustom...)

		default:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, args[0])
		}

		return nil
	}

	return fmt.Errorf(ErrFieldPermission, field, "filter")
}

// addSortCondition adds an ORDER BY condition with the given controller params.
// Error will return if the field is not allowed to sort or does not exist.
func addSortCondition(g *grid, params string, c condition.Condition) error {
	sortFields := strings.Split(params, conditionSortSeparator)
	var orderFields []string

	// skip if there are arguments.
	if len(sortFields) == 1 && sortFields[0] == "" {
		return nil
	}

	// checking if the field is allowed for sorting
	for _, f := range sortFields {
		prefix := ""
		if strings.HasPrefix(f, "-") {
			f = f[1:]
			prefix = "-"
		}
		if gridField := g.Field(f); gridField.error == nil && gridField.sortAble {
			orderFields = append(orderFields, prefix+gridField.sortField)
		} else {
			return fmt.Errorf(ErrFieldPermission, f, conditionSortKey)
		}
	}

	// adding order
	c.SetOrder(orderFields...)

	return nil
}

// escape is needed to escape the % and _ for sql.
// at the moment it will only be prefixed with a backslash.
// TODO: the source must provide an escape function.
func escape(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "%", "\\%"), "_", "\\_")
}
