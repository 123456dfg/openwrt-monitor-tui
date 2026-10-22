package kit

import (
	"reflect"
	"testing"
)

// Every visible string must be translated: an unset field would render as a
// blank label or an empty hint bar instead of failing loudly.
func TestTranslationsComplete(t *testing.T) {
	for name, tr := range map[string]Strings{"en": en, "zh": zh} {
		v := reflect.ValueOf(tr)
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			switch f.Kind() {
			case reflect.String:
				if f.String() == "" {
					t.Errorf("%s: %s is empty", name, v.Type().Field(i).Name)
				}
			case reflect.Array:
				if f.Type().Elem().Kind() != reflect.String {
					continue
				}
				for j := 0; j < f.Len(); j++ {
					if f.Index(j).String() == "" {
						t.Errorf("%s: %s[%d] is empty", name, v.Type().Field(i).Name, j)
					}
				}
			}
		}
	}
}
