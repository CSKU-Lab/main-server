package sqlx

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func getUpdateFields(s any) string {
	typ := reflect.TypeOf(s)

	if typ.Kind() != reflect.Pointer {
		if typ.Elem().Kind() != reflect.Struct {
			log.Fatalln("The argument that you passed is not a struct")
		}
		log.Fatalln("The argument that you passed is not a pointer")
	}

	fields := getAllStructFields(s)

	return strings.Join(fields, ",")
}

func getAllStructFields(s any) []string {
	typ := reflect.TypeOf(s)
	val := reflect.ValueOf(s)
	keys := make([]string, 0)

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := range val.NumField() {
		field := typ.Field(i)
		value := val.Field(i)

		if field.Anonymous {
			keys = append(keys, getAllStructFields(value.Interface())...)
		}

		if !value.IsValid() {
			continue
		}
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				continue
			}
		} else if value.IsZero() {
			continue
		}

		key := field.Tag.Get("db")

		if key == "" {
			continue
		}

		keys = append(keys, fmt.Sprintf("%s = :%s", key, key))
	}

	return keys
}
