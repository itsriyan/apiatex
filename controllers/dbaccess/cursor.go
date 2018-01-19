package dbaccess

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Filter struct {
	Field string
	//Op is binary comparison :
	//	>  greater than
	//	>= greater than or equal
	//	< less than
	//	<= less than or equal
	//	= equal
	//	!= not equal
	Op    string
	Value interface{}
}

func (f Filter) IsValidOp() error {
	if f.Op == "like" || f.Op == ">" || f.Op == ">=" || f.Op == "<" || f.Op == "<=" || f.Op == "=" || f.Op == "!=" {
		return nil
	}
	return errors.New("invalid cursor op")
}

type Cursor struct {
	Fields     []string
	Filters    []Filter
	OrderBy    []string
	Limit      int
	Descending bool
	//LastArgs store last value of the item to use for the filter on the next fetch.
	LastArgs []interface{}
}

var sep = []byte("\n")

func (c Cursor) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(strings.Join(c.Fields, ","))
	buf.Write(sep)
	for _, filter := range c.Filters {
		fmt.Fprintf(buf, "%s,%s,%v;", filter.Field, filter.Op, filter.Value)
	}
	buf.Write(sep)
	buf.WriteString(strings.Join(c.OrderBy, ","))
	buf.Write(sep)
	fmt.Fprintf(buf, "%d", c.Limit)
	buf.Write(sep)
	fmt.Fprintf(buf, "%t", c.Descending)
	buf.Write(sep)
	buf.WriteString(siToString(c.LastArgs))
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

//MarshalJSON implement json marshaller.
func (c Cursor) MarshalJSON() (data []byte, err error) {
	data = []byte(fmt.Sprintf("%q", c.String()))
	return
}

func (c *Cursor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	cs, err := Decode(s)
	if err != nil {
		return err
	}
	*c = cs
	return nil
}

func Decode(s string) (Cursor, error) {
	c := Cursor{}
	if len(s) == 0 {
		return c, nil
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return c, err
	}
	ds := bytes.Split(data, sep)
	if len(ds) == 0 {
		return c, nil
	}
	c.Fields = decodeToSS(ds[0])
	if len(ds) > 1 {
		c.Filters = decodeFilters(ds[1])
	}
	if len(ds) > 2 {
		c.OrderBy = decodeToSS(ds[2])
	}
	if len(ds) > 3 {
		limit, err := strconv.Atoi(string(ds[3]))
		if err != nil {
			return c, err
		}
		c.Limit = limit
	}
	if len(ds) > 4 {
		if string(ds[4]) == "true" {
			c.Descending = true
		} else {
			c.Descending = false
		}
	}
	if len(ds) > 5 {
		args := decodeToSS(ds[5])
		if len(args) != 0 {
			if len(c.LastArgs) != len(args) {
				c.LastArgs = make([]interface{}, len(args))
			}
			for i, v := range args {
				c.LastArgs[i] = v
			}
		}
	}
	return c, nil
}

func siToString(args []interface{}) string {
	var s []string
	for _, v := range args {
		switch x := v.(type) {
		case int16:
			s = append(s, strconv.Itoa(int(x)))
		case int:
			s = append(s, strconv.Itoa(x))
		case string:
			s = append(s, x)
		case *string:
			s = append(s, *x)
		case float64:
			s = append(s, strconv.FormatFloat(x, 'f', -1, 64))
		case fmt.Stringer:
			s = append(s, x.String())
		case *int:
			s = append(s, strconv.Itoa(*x))
		case *int16:
			s = append(s, strconv.Itoa(int(*x)))
		case *float64:
			s = append(s, strconv.FormatFloat(*x, 'f', -1, 64))
		default:
			panic("cursor: can not convert interface to string")
		}
	}
	return strings.Join(s, ",")
}

func decodeToSS(data []byte) []string {
	var result []string
	if len(data) == 0 {
		return result
	}
	result = strings.Split(string(data), ",")
	return result
}

func decodeFilters(data []byte) []Filter {
	var filters []Filter
	if len(data) == 0 {
		return filters
	}
	ds := bytes.Split(data, []byte(";"))
	for _, v := range ds {
		if len(v) != 0 {
			vs := bytes.Split(v, []byte(","))
			filters = append(filters,
				Filter{
					Field: string(vs[0]),
					Op:    string(vs[1]),
					Value: string(vs[2]),
				})
		}
	}
	return filters
}
