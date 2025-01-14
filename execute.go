package luatemplate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"text/template"

	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

var ErrInvalidTemplate = errors.New("invalid lua template")

// New like [text/template.New]
func New(name string) *LuaTemplate {
	return &LuaTemplate{
		name: name,
	}
}

type LuaTemplate struct {
	name   string
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

// Parse like [text/template.Template] Parse.
func (lt *LuaTemplate) Parse(source string) (*LuaTemplate, error) {
	kit := luakit.Apply("vela", func(p lua.Preloader) {
		p.SetGlobal("gee", lua.NewGeneric(lt))
	})

	state := kit.NewState(context.Background())
	defer state.Close()

	if err := state.DoString(source); err != nil {
		return nil, err
	}
	if lt.exdata == nil {
		return nil, ErrInvalidTemplate
	}
	lt.source = source

	return lt, nil
}

// Execute like [text/template.Template] Execute.
func (lt *LuaTemplate) Execute(w io.Writer, data any) error {
	switch v := data.(type) {
	case json.RawMessage:
		param := make(map[string]any, 16)
		if err := json.Unmarshal(v, &param); err != nil {
			return err
		}
		data = param
	}

	tmpl, err := template.New(lt.name).
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

	return nil
}
