// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"encoding/csv"
	"fmt"
	"github.com/patrickascher/gofer/controller/context"
	"reflect"
	"time"
)

func init() {
	_ = context.RegisterRenderer("gridCsv", newCsv)
}

// New satisfies the config.provider interface.
func newCsv() (context.Renderer, error) {
	return &csvWriter{}, nil
}

type csvWriter struct {
}

func (cw *csvWriter) Name() string {
	return "CSV"
}

func (cw *csvWriter) Icon() string {
	return "mdi-file-delimited-outline"
}

func (cw *csvWriter) Error(r *context.Response, code int, err error) error {
	r.Writer().WriteHeader(code)
	_, err = r.Writer().Write([]byte(err.Error()))
	return err
}

func (cw *csvWriter) Write(r *context.Response) error {

	// Filename
	filename := "export"
	if r.Value(FILENAME) != nil {
		filename = r.Value(FILENAME).(string)
	}

	r.Writer().Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Writer().Header().Set("Content-Disposition", "attachment; filename=\""+filename+".csv\"")

	w := csv.NewWriter(r.Writer())
	w.Comma = 59 //;

	// UTF-8 BOM for Excel
	bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
	err := w.Write([]string{string(bomUtf8[:])})
	if err != nil {
		return err
	}

	var header []Field
	for _, h := range r.Value("head").([]Field) {
		if h.Removed() {
			continue
		}
		header = append(header, h)
	}

	data := r.Value("data")

	//header
	var headString []string
	for _, head := range header {
		headString = append(headString, head.Title())
	}
	if err := w.Write(headString); err != nil {
		return err
	}

	// adding body
	rData := reflect.ValueOf(data)
	for i := 0; i < rData.Len(); i++ {
		var body []string

		for _, head := range header {
			if rData.Index(i).Type().Kind().String() == "struct" {
				if rData.Index(i).FieldByName(head.name).Type().String() == "query.NullString" {
					body = append(body, fmt.Sprint(rData.Index(i).FieldByName(head.name).FieldByName("String").Interface()))
				} else {
					body = append(body, fmt.Sprint(rData.Index(i).FieldByName(head.name).Interface()))
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
								return err
							}
							body = append(body, t.Format(dateFormat[0:10]))
							continue
						case "DateTime":
							t, err := time.Parse("2006-01-02 15:04", date[0:16])
							if err != nil {
								return err
							}
							body = append(body, t.Format(dateFormat[0:16]))
							continue
						}
					} else {
						body = append(body, "")
					}
					continue
				}
				body = append(body, fmt.Sprint(reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head.name)).Interface()))
			}
		}

		if err := w.Write(body); err != nil {
			return err
		}
	}

	w.Flush()

	return nil
}
