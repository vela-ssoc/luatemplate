package luatemplate

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var MyFuncMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"lua":      Lua,
	"join":     Join,      // {{ join  ["123" , "456"] "," }}
	"table":    ViewTable, // {{ table  data "id:%v,value:%v" "id" "value" }} or {{ table data "key:%s value:%s" "key" "value" }}
	"array":    ViewArray,
	"indent":   ViewIndent,
	"checkbox": ViewCheckBox,
}

func Lua(v string) string {
	// todo
	return v
}

func ViewTable(data []interface{}, format string, keys ...string) (string, error) {
	n := len(data)
	if n == 0 {
		return "", nil
	}

	buff := new(bytes.Buffer)
	render := func(tab map[string]interface{}) {
		var val []interface{}
		for _, key := range keys {
			item, ok := tab[key]
			if !ok {
				continue
			}
			val = append(val, item)
		}

		buff.WriteString(fmt.Sprintf(format, val...))
	}

	for i := 0; i < n; i++ {
		switch elem := data[i].(type) {
		case map[string]interface{}:
			render(elem)
		case string, int, float64, bool:
			tab := map[string]interface{}{
				"id":    i + 1,
				"value": elem,
			}
			render(tab)
		default:
			return "", fmt.Errorf("not suppert %T", elem)
		}
	}

	return buff.String(), nil
}

func Join(arr []interface{}, sep string) (string, error) {
	var buff bytes.Buffer
	n := len(arr)
	for i := 0; i < n; i++ {
		switch el := arr[i].(type) {
		case int, float64, float32, bool:
			buff.WriteString(fmt.Sprintf("%v", el))
		default:
			buff.WriteString(fmt.Sprintf("\"%v\"", el))
		}

		if i != n-1 {
			buff.WriteString(sep)
		}
	}

	return buff.String(), nil
}

func ViewArray(arr []interface{}, format string, keys ...string) (string, error) {
	n := len(arr)
	if n == 0 {
		return "", nil
	}

	f := len(keys)
	buff := new(bytes.Buffer)
	if 0 == f {
		for i := 0; i < n; i++ {
			buff.WriteString(fmt.Sprintf(format, arr[i]))
		}
		return buff.String(), nil
	}

	fn := func(tab map[string]interface{}) string {
		var val []interface{}
		for _, key := range keys {
			item, ok := tab[key]
			if !ok {
				continue
			}
			val = append(val, item)
		}

		return fmt.Sprintf(format, val...)
	}

	for i := 0; i < n; i++ {
		elem, ok := arr[i].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("not map")
		}
		buff.WriteString(fn(elem))
	}

	return buff.String(), nil
}

func ViewIndent(indent int, v string) string {
	data := strings.Replace(v, "\n", "\n"+strings.Repeat(" ", indent), -1)
	return data
}

func ViewCheckBox(v interface{}, label string) (string, error) {
	switch el := v.(type) {
	case bool:
		if el {
			return label, nil
		}

		return "", nil
	case string:
		if len(el) > 0 {
			return label, nil
		}
		return "", nil
	}

	return "", fmt.Errorf("not checkbox element value")
}
