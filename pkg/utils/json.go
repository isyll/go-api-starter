package utils

import (
	"reflect"
	"strings"
)

func GetJSONFieldName(typ reflect.Type, structFieldName string) string {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	parts := strings.Split(structFieldName, ".")

	currentType := typ
	jsonPath := make([]string, 0, len(parts))

	for _, part := range parts {
		field, found := currentType.FieldByName(part)
		if !found {
			jsonPath = append(jsonPath, part)
			continue
		}

		tag := field.Tag.Get("json")
		switch tag {
		case "-":
			jsonPath = append(jsonPath, part)
		case "":
			jsonPath = append(jsonPath, part)
		default:
			name := strings.Split(tag, ",")[0]
			jsonPath = append(jsonPath, name)
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		currentType = fieldType
	}

	return strings.Join(jsonPath, ".")
}
