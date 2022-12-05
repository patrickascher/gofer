// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"reflect"
	"time"

	"github.com/patrickascher/gofer/controller/context"
	"github.com/xuri/excelize/v2"
)

func init() {
	_ = context.RegisterRenderer("gridExcel", newXls)
}

// New satisfies the config.provider interface.
func newXls() (context.Renderer, error) {
	return &newExcel{}, nil
}

type newExcel struct {
}

func (cw *newExcel) Name() string {
	return "Excel"
}

func (cw *newExcel) Icon() string {
	return "mdi-microsoft-excel"
}

func (cw *newExcel) Error(r *context.Response, code int, err error) error {
	r.Writer().WriteHeader(code)
	_, err = r.Writer().Write([]byte(err.Error()))
	return err
}

func (cw *newExcel) Write(r *context.Response) error {

	// Filename
	filename := "export"
	if r.Value(FILENAME) != nil {
		filename = r.Value(FILENAME).(string)
	}

	r.Writer().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet; charset=utf-8")
	r.Writer().Header().Set("Content-Disposition", "attachment; filename=\""+filename+".xlsx\"")

	// sheet name
	sheetName := "Sheet1"
	f := excelize.NewFile()

	//header
	var header []Field
	for _, h := range r.Value("head").([]Field) {
		if h.Removed() {
			continue
		}
		header = append(header, h)
	}
	for i, head := range header {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return err
		}
		err = f.SetCellValue(sheetName, cell, head.Title())
		if err != nil {
			return err
		}
	}

	// adding body
	data := r.Value("data")
	rData := reflect.ValueOf(data)
	for i := 0; i < rData.Len(); i++ {

		for h, head := range header {
			cell, err := excelize.CoordinatesToCellName(h+1, i+2)
			if err != nil {
				return err
			}

			if rData.Index(i).Type().Kind().String() == "struct" {
				if rData.Index(i).FieldByName(head.name).Type().String() == "query.NullString" {
					err = f.SetCellValue(sheetName, cell, fmt.Sprint(rData.Index(i).FieldByName(head.name).FieldByName("String").Interface()))
					if err != nil {
						return err
					}
				} else {
					err = f.SetCellValue(sheetName, cell, fmt.Sprint(rData.Index(i).FieldByName(head.name).Interface()))
					if err != nil {
						return err
					}
				}
			} else {
				// for date string values - recast it.
				if head.Type() == "Date" || head.Type() == "DateTime" {
					dateFormat := "2006-01-02 15:04"
					if d := r.Value(DATEFORMAT); d != nil {
						dateFormat = d.(string)
					}
					date := fmt.Sprint(reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head.name)).Interface())
					if date != "" && date != "<nil>" {
						switch head.Type() {
						case "Date":
							t, err := time.Parse("2006-01-02", date[0:10])
							if err != nil {
								fmt.Println(err)
								return err
							}
							err = f.SetCellValue(sheetName, cell, t.Format(dateFormat[0:10]))
							if err != nil {
								return err
							}
							continue
						case "DateTime":
							t, err := time.Parse("2006-01-02 15:04", date[0:16])
							if err != nil {
								fmt.Println(err)
								return err
							}
							err = f.SetCellValue(sheetName, cell, t.Format(dateFormat[0:16]))
							if err != nil {
								return err
							}
							continue
						}
					} else {
						err = f.SetCellValue(sheetName, cell, "")
						if err != nil {
							return err
						}
					}
					continue
				}
				err = f.SetCellValue(sheetName, cell, fmt.Sprint(reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head.name)).Interface()))
				if err != nil {
					return err
				}
			}
		}
	}

	return f.Write(r.Writer())
}
