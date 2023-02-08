package determinator

import (
	"fmt"
	"reflect"

	"github.com/deliveroo/determinator-go/models"
	validator "github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

// Actors represents the context's actors.
type Actors struct {
	Request    *models.Request
	Customer   *models.Customer
	Restaurant *models.Restaurant
	Rider      *models.Rider
}

// ToParams returns a map with dotted
func (a *Actors) ToParams() (map[string]string, error) {
	res := map[string]string{}
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return res, err
	}
	toParams(a, &res, "")
	return res, nil
}

func toParams(model interface{}, res *map[string]string, prefix string) {
	modelReflect := reflect.ValueOf(model)

	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}

	modelRefType := modelReflect.Type()
	fieldsCount := modelReflect.NumField()

	for i := 0; i < fieldsCount; i++ {
		structField := modelRefType.Field(i)
		fieldName := strcase.ToSnake(structField.Name)
		if param, ok := structField.Tag.Lookup("param"); ok {
			fieldName = param
		}

		field := modelReflect.Field(i)

		if field.IsZero() {
			continue
		}

		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			toParams(field.Interface(), res, prefix+fieldName+".")

		default:
			fieldData := field.Interface()
			(*res)[prefix+fieldName] = fmt.Sprintf("%v", fieldData)
		}
	}
}
