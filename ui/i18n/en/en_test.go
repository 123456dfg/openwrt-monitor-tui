package en

import (
	"reflect"
	"testing"
)

// Every visible string must be translated: an unset field would render as a
// blank label or an empty hint bar instead of failing loudly.
func TestTableComplete(t *testing.T) {
	v := reflect.ValueOf(T)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			if f.String() == "" {
				t.Errorf("%s is empty", v.Type().Field(i).Name)
			}
		case reflect.Array:
			if f.Type().Elem().Kind() != reflect.String {
				continue
			}
			for j := 0; j < f.Len(); j++ {
				if f.Index(j).String() == "" {
					t.Errorf("%s[%d] is empty", v.Type().Field(i).Name, j)
				}
			}
		}
	}
}
