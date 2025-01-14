package luatemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/luakit"
	"io"
	"text/template"

	"github.com/vela-public/onekit/lua"
)

func New() *LuaTemplate {
	return &LuaTemplate{}
}

type LuaTemplate struct {
	source string
	exdata *Template
}

func (lt *LuaTemplate) NewTemplateL(L *lua.LState) int {
	lt.exdata = &Template{
		Info: &Info{Version: "0.0.0", Author: "vela"},
	}
	tab := L.CheckTable(1)
	err := luakit.TableTo(L, tab, lt.exdata)
	if err != nil {
		L.RaiseError("%v", err)
		return 0
	}
	L.Push(lt.exdata)
	return 1

}

func (lt *LuaTemplate) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "param":
		return lua.NewExport("lua.param.export", lua.WithIndex(ParamIndexL))
	case "template":
		return lua.NewFunction(lt.NewTemplateL)

	}
	return lua.LNil
}

func (lt *LuaTemplate) Parse(source string) (*LuaTemplate, error) {
	kit := luakit.Apply("vela", func(p lua.Preloader) {
		p.SetGlobal("gee", lua.NewGeneric[*LuaTemplate](lt))
	})

	state := kit.NewState(context.Background(), func(opt *lua.Options) {
		//todo
	})
	defer state.Close()

	if err := state.DoString(source); err != nil {
		return nil, err
	}
	lt.source = source
	if lt.exdata == nil {
		return nil, fmt.Errorf("not found template")
	}
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
		Parse(lt.exdata.Text)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func (lt *LuaTemplate) ParamJSON() json.RawMessage {
	if bs, err := lt.exdata.ParamJSON(false); err == nil {
		return bs
	}

	return json.RawMessage("null")
}
