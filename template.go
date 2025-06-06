package luatemplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/tidwall/pretty"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
)

type Info struct {
	File    string
	Version string
	Author  string
	CTime   int64
}

type Template struct {
	Info   *Info
	Param  []*Param
	Buffer bytes.Buffer
	Text   string
}

func (t *Template) String() string                         { return "" }
func (t *Template) Type() lua.LValueType                   { return lua.LTObject }
func (t *Template) AssertFloat64() (float64, bool)         { return 0, false }
func (t *Template) AssertString() (string, bool)           { return "", false }
func (t *Template) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (t *Template) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (t *Template) ParamJSON(p bool) ([]byte, error) {
	if t.Param == nil {
		return nil, fmt.Errorf("not found")
	}

	enc := jsonkit.NewJson()
	enc.Arr("")
	for _, item := range t.Param {
		enc.Copy(item.Marshal())
		enc.Char(',')
	}
	enc.End("]")

	chunk := enc.Bytes()
	if p {
		return pretty.Pretty(chunk), nil
	}
	return chunk, nil
}

func (t *Template) Schema() ([]byte, error) {
	return nil, nil
}

func (t *Template) GenLuaCode(L *lua.LState) int {
	return 0
}

func (t *Template) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "gen_lua_code":
		return lua.NewFunction(t.GenLuaCode)
	}
	return nil
}

func (t *Template) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "param":
		tab := lua.CheckTable(L, val)
		t.Param = Tab2ParamTab(L, tab)

	case "text":
		t.Text = val.String()
	}
}

func (t *Template) Chunk() []byte {
	return t.Buffer.Bytes()
}

func (t *Template) check(v map[string]interface{}) error {
	return nil
}

func (t *Template) Bind(data []byte, v map[string]interface{}) error {
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	return t.check(v)
}

func (t *Template) BindJSON(data []byte) error {
	v := make(map[string]interface{})
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	return t.Gen(v)
}

func (t *Template) Gen(v interface{}) error {
	tp, err := template.New("lua").Funcs(MyFuncMap).Parse(t.Text)
	if err != nil {
		return err
	}

	t.Buffer.Reset()

	err = tp.Execute(&t.Buffer, v)
	if err != nil {
		return err
	}

	return nil
}

func (t *Template) copy(v2 *Template) {
	// copy info
	t.Info.File = v2.Info.File
	t.Info.Version = v2.Info.Version
	t.Info.Author = v2.Info.Author

	// copy param
	t.Param = v2.Param
	t.Text = v2.Text
	t.Buffer = v2.Buffer
}

// NewTemplateL 是一个用于生成TemplateGenLua的函数
// L 是 lua.LState 类型的参数，表示 Lua 虚拟机的状态
// 函数返回一个整数值，表示生成的TemplateGenLua在Lua栈中的位置
func NewTemplateL(L *lua.LState) int {
	t := &Template{
		Info: &Info{Version: "0.0.0", Author: "vela"},
	}
	lua.SetExdata2(L, t)

	tab := L.CheckTable(1)
	tab.Range(func(key string, val lua.LValue) {
		switch key {
		case "version":
			t.Info.Version = val.String()
		case "auth", "author":
			t.Info.Author = val.String()
		}
	})

	L.Push(t)
	return 1
}
