package scanstruct

import (
	"database/sql"
	"reflect"
	"strings"
)

func ScanStruct(targetStruct interface{}, r *sql.Rows) (err error) {
	t := reflect.TypeOf(targetStruct)
	if t.Kind() != reflect.Ptr {
		panic("Not a pointer")
	}
	columns, _ := r.Columns()
	val := reflect.ValueOf(targetStruct)
	valVal := val.Elem()
	var pointers []interface{}
	for _, column := range columns {
		f := valVal.FieldByNameFunc(func(fieldName string) bool {
			return strings.ToLower(column) == strings.ToLower(fieldName)
		})
		if !f.IsValid() {
			panic("Have not gotten that far yet")
		} else {
			pointers = append(pointers, f.Addr().Interface())
		}
	}
	err = r.Scan(pointers...)
	return
}
