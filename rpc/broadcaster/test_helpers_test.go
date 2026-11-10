package broadcaster

import "reflect"

func field(v reflect.Value, name string) reflect.Value {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v.FieldByName(name)
}
