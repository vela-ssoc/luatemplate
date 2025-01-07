package luatemplate

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/vela-public/onekit/lua"
)

func New() *LuaTemplate {
	return &LuaTemplate{}
}

type LuaTemplate struct {
	source string
	tmpl   *Template
}

func (lt *LuaTemplate) Parse(source string) (*LuaTemplate, error) {
	state := lua.NewState()
	defer state.Close()

	state.SetGlobal("gee", GenLuaCodeL())
	if err := state.DoString(source); err != nil {
		return nil, err
	}

	tmpl, _ := state.A().(*Template)
	if tmpl == nil {
		return nil, fmt.Errorf("state.A() must be of type *Template")
	}

	lt.source = source
	lt.tmpl = tmpl

	return lt, nil
}

// Execute Like [text/template.Template]
func (lt *LuaTemplate) Execute(w io.Writer, data any) error {
	switch v := data.(type) {
	case json.RawMessage:
		param := make(map[string]any, 16)
		if err := json.Unmarshal(v, &param); err != nil {
			return err
		}
		data = param
	}

	tmpl, err := template.New("lua").
		Funcs(MyFuncMap).
		Parse(lt.tmpl.Text)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func (lt *LuaTemplate) ParamJSON() json.RawMessage {
	if bs, err := lt.tmpl.ParamJSON(false); err == nil {
		return bs
	}

	return json.RawMessage("null")
}
