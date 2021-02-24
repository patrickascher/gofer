package grid

import (
	"encoding/csv"
	"fmt"
	"github.com/patrickascher/gofer/controller/context"
	"reflect"
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

	// TODO define separator
	// TODO define CRLF

	r.Writer().Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Writer().Header().Set("Content-Disposition", "attachment; filename=\"export.csv\"")

	w := csv.NewWriter(r.Writer())
	w.Comma = 59 //;

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
