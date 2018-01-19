package dbaccess

import (
	"reflect"
	"testing"
)

func TestCursor(t *testing.T) {
	testData := []Cursor{
		{
			Fields:  []string{"id", "name"},
			OrderBy: []string{"name"},
			Limit:   50,
		},
		{
			Fields:   []string{"id", "name"},
			OrderBy:  []string{"name"},
			Limit:    50,
			LastArgs: []interface{}{"PCS", "1000"},
		},
		{
			Fields: []string{"id", "name"},
			Filters: []Filter{
				{Field: "code", Op: "=", Value: "C01"},
			},
			OrderBy:  []string{"name"},
			Limit:    50,
			LastArgs: []interface{}{"PCS", "1000"},
		},
	}
	for i, c := range testData {
		enc := c.String()
		newC, err := Decode(enc)
		if err != nil {
			t.Errorf("tc:%d got err:%v want:nil", i, err)
		}
		if !reflect.DeepEqual(c, newC) {
			t.Errorf("tc:%d got:%#v want:%#v", i, newC, c)
		}
	}
}
