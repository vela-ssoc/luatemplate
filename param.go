package luatemplate

import (
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
)

type ParamKV struct {
	key string
	val interface{}
}

type OptionKV struct {
	Value any    `json:"value"`
	Label string `json:"label"`
}

type Param struct {
	Name    string
	Label   string
	Desc    string
	Mime    string
	Mask    []ParamKV
	Style   []ParamKV // {span=18 , label="inner" , labelLayout="x" , collapse=1}
	Array   []*Param
	RawData []byte
}

func (p *Param) Hijack(fsm *lua.CallFrameFSM) bool { return false }

func (p *Param) Marshal() []byte {
	enc := jsonkit.NewJson()
	enc.Tab("")
	enc.KV("name", p.Name)
	enc.KV("mime", p.Mime)
	enc.KV("label", p.Label)
	enc.KV("desc", p.Desc)

	enc.Tab("style")
	for _, kv := range p.Style {
		enc.KV(kv.key, kv.val)
	}
	enc.End("},")

	switch p.Mime {
	case "array", "map":
		enc.Arr(p.Mime)
		for _, item := range p.Array {
			enc.Copy(item.Marshal())
			enc.Char(',')
		}
		enc.End("],")

		enc.Tab("mask")
		if len(p.RawData) > 0 {
			enc.Raw("default", p.RawData)
		}

		for _, kv := range p.Mask {
			enc.KV(kv.key, kv.val)
		}

		enc.End("},")

	default:
		enc.Tab("mask")
		for _, kv := range p.Mask {
			enc.KV(kv.key, kv.val)
		}
		enc.End("},")
	}

	enc.End("}")
	return enc.Bytes()
}

func (p *Param) have(key string) bool {
	n := len(p.Mask)
	if n == 0 {
		return false
	}
	for i := 0; i < n; i++ {
		if p.Mask[i].key == key {
			return true
		}
	}
	return false
}

func (p *Param) SetKV(key string, val interface{}) {
	p.Mask = SetKV(p.Mask, key, val)
}

func (p *Param) String() string                         { return "" }
func (p *Param) Type() lua.LValueType                   { return lua.LTObject }
func (p *Param) AssertFloat64() (float64, bool)         { return 0, false }
func (p *Param) AssertString() (string, bool)           { return "", false }
func (p *Param) AssertFunction() (*lua.LFunction, bool) { return p.ToLFunc(), true }

func (p *Param) ToLFunc() *lua.LFunction {
	return lua.NewFunction(func(L *lua.LState) int {
		// name
		p.Name = L.IsString(1)

		L.Push(p)
		return 1
	})
}

func (p *Param) sizeL(L *lua.LState) int {
	n := L.IsInt(1)
	if n > 0 {
		p.SetKV("size", n)
	}

	L.Push(p)
	return 1
}

func (p *Param) regexL(L *lua.LState) int {
	r := L.IsString(1)
	if len(r) > 0 {
		p.SetKV("regex", r)
	}
	L.Push(p)
	return 1
}

func (p *Param) mustL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		L.Push(p)
		return 1
	}

	var result []interface{}
	var last_t lua.LValueType

	update := func(lv lua.LValue) {
		vt := lv.Type()
		if last_t != vt {
			L.RaiseError("must Too many types %s %s", last_t.String(), lv.Type().String())
			return
		}

		switch vt {
		case lua.LTString:
			result = append(result, lv.String())
		case lua.LTNumber:
			result = append(result, float64(lv.(lua.LNumber)))
		case lua.LTInt:
			result = append(result, int(lv.(lua.LInt)))
		case lua.LTBool:
			result = append(result, bool(lv.(lua.LBool)))
		default:
			result = append(result, lv.String())
		}
	}

	for i := 1; i <= n; i++ {
		lv := L.Get(i)
		if i == 1 {
			last_t = lv.Type()
		}
		update(lv)
	}

	p.SetKV("must", result)
	L.Push(p)
	return 1
}

func (p *Param) minL(L *lua.LState) int {
	v := L.IsInt(1)
	p.SetKV("min", v)
	L.Push(p)
	return 1
}

func (p *Param) maxL(L *lua.LState) int {
	v := L.IsInt(1)
	p.SetKV("max", v)
	L.Push(p)
	return 1
}

func (p *Param) DefaultLinesL(L *lua.LState) int {
	top := L.GetTop()
	if top == 0 {
		L.Push(p)
		return 1
	}

	data := lua.L2SS(L)

	p.SetKV("default", data)
	L.Push(p)
	return 1
}

func (p *Param) defaultL(L *lua.LState) int {
	top := L.GetTop()
	if top == 0 {
		L.Push(p)
		return 1
	}

	lv := L.Get(1)
	switch lv.Type() {
	case lua.LTString:
		arr := NewDefault[string](top)
		for i := 1; i <= top; i++ {
			arr[i-1] = L.Get(i).String()
		}

		if top == 1 {
			p.SetKV("default", arr[0])
		} else {
			p.SetKV("default", arr)
		}

	case lua.LTInt:
		arr := NewDefault[int](top)
		for i := 1; i <= top; i++ {
			arr[i-1] = int(L.Get(i).(lua.LInt))
		}

		if top == 1 {
			p.SetKV("default", arr[0])
		} else {
			p.SetKV("default", arr)
		}

	case lua.LTBool:
		arr := NewDefault[bool](top)
		for i := 1; i <= top; i++ {
			arr[i-1] = bool(L.Get(i).(lua.LBool))
		}

		if top == 1 {
			p.SetKV("default", arr[0])
		} else {
			p.SetKV("default", arr)
		}
	case lua.LTNumber:
		arr := NewDefault[float64](top)
		for i := 1; i <= top; i++ {
			arr[i-1] = float64(L.Get(i).(lua.LNumber))
		}

		if top == 1 {
			p.SetKV("default", arr[0])
		} else {
			p.SetKV("default", arr)
		}
	}

	L.Push(p)
	return 1
}

func (p *Param) tableL(L *lua.LState) int {
	tab := L.CheckTable(1)
	arr := tab.Array()
	if len(arr) == 0 {
		L.RaiseError("array is empty")
		return 0
	}

	for _, item := range arr {
		v, ok := item.(*Param)
		if ok {
			p.Array = append(p.Array, v)
		}
	}
	L.Push(p)
	return 1
}

func (p *Param) descL(L *lua.LState) int {
	desc := L.IsString(1)
	if len(desc) > 0 {
		p.Desc = desc
	}
	L.Push(p)
	return 1
}

func (p *Param) apiL(L *lua.LState) int {
	cmd := L.IsString(1)
	if len(cmd) > 0 {
		p.Mask = SetKV(p.Mask, "api", cmd)
	}
	L.Push(p)
	return 1
}

func (p *Param) EnvWhereL(L *lua.LState) int {
	cmd := L.IsString(1)
	if len(cmd) > 0 {
		p.Mask = SetKV(p.Mask, "where", cmd)
	}
	L.Push(p)
	return 1
}

func (p *Param) bindL(L *lua.LState) int {
	cmd := L.IsString(1)
	if len(cmd) > 0 {
		p.Mask = SetKV(p.Mask, "bind", cmd)
	}
	L.Push(p)
	return 1
}

func (p *Param) modeL(L *lua.LState) int {
	cmd := L.IsString(1)
	if len(cmd) > 0 {
		p.Mask = SetKV(p.Mask, "mode", cmd)
	}
	L.Push(p)
	return 1
}

func (p *Param) condL(L *lua.LState) int {
	cnd := L.IsString(1)
	if len(cnd) > 0 {
		p.Mask = SetKV(p.Mask, "cond", cnd)
	}
	L.Push(p)
	return 1
}

func (p *Param) labeLL(L *lua.LState) int {
	label := L.IsString(1)
	if len(label) > 0 {
		p.Label = label
	}
	L.Push(p)
	return 1
}

func (p *Param) styleL(L *lua.LState) int {
	tab := L.CheckTable(1)

	tab.Range(func(key string, val lua.LValue) {
		v, ok := V2I(val)
		if ok {
			p.Style = SetKV(p.Style, key, v)
		}
	})

	L.Push(p)
	return 1
}

func (p *Param) arrL(L *lua.LState) int {
	tab := L.CheckTable(1)
	if !tab.IsArray() {
		L.RaiseError("not found array default value")
		return 0
	}

	arr := tab.Array()

	enc := jsonkit.NewJson()
	enc.Arr("")
	for _, item := range arr {
		switch item.Type() {
		case lua.LTInt:
		case lua.LTNumber:
		case lua.LTTable:
			enc.Tab("")
			tv, _ := item.(*lua.LTable)

			tv.Range(func(key string, val lua.LValue) {
				v, ok := V2I(val)
				if ok {
					enc.KV(key, v)
				}
			})

			enc.End("},")
		}
	}
	enc.End("]")

	p.RawData = enc.Bytes()
	L.Push(p)
	return 1
}

func (p *Param) optionL(L *lua.LState) int {
	value := L.Get(1)
	label := L.IsString(2)

	kv := OptionKV{
		Label: label,
	}

	switch value.Type() {
	case lua.LTInt:
		kv.Value = int(value.(lua.LInt))
	case lua.LTNumber:
		kv.Value = float64(value.(lua.LNumber))
	case lua.LTBool:
		kv.Value = bool(value.(lua.LBool))
	case lua.LTString:
		kv.Value = value.String()
	default:
		kv.Value = value.String()
	}

	data, ok := Pickup[[]OptionKV](p.Mask, "options")
	if !ok {
		p.SetKV("options", []OptionKV{kv})
		L.Push(p)
		return 1
	}

	for _, item := range data {
		if item.Value == kv.Value {
			L.Push(p)
			return 1
		}
	}

	data = append(data, kv)
	p.SetKV("options", data)
	L.Push(p)
	return 1
}

func (p *Param) propertyL(L *lua.LState, key string) lua.LValue {
	switch key {
	case "size":
		return lua.NewFunction(p.sizeL)
	case "min":
		return lua.NewFunction(p.minL)
	case "max":
		return lua.NewFunction(p.maxL)
	case "must":
		return lua.NewFunction(p.mustL)
	case "regex":
		return lua.NewFunction(p.regexL)
	case "default":
		return lua.NewFunction(p.defaultL)
	case "desc":
		return lua.NewFunction(p.descL)
	case "label":
		return lua.NewFunction(p.labeLL)
	case "style":
		return lua.NewFunction(p.styleL)
	case "cond":
		return lua.NewFunction(p.condL)
	case "api":
		return lua.NewFunction(p.apiL)
	case "mode":
		return lua.NewFunction(p.modeL)
	case "option":
		return lua.NewFunction(p.optionL)
	default:
		L.RaiseError("not found property %s", key)
	}

	return p
}

func (p *Param) Index(L *lua.LState, key string) lua.LValue {
	switch p.Mime {
	case "array", "map":
		switch key {
		case "table":
			return lua.NewFunction(p.tableL)
		case "label":
			return lua.NewFunction(p.labeLL)
		case "size":
			return lua.NewFunction(p.sizeL)
		case "default":
			return lua.NewFunction(p.arrL)
		case "style":
			return lua.NewFunction(p.styleL)
		case "desc":
			return lua.NewFunction(p.descL)
		case "cond":
			return lua.NewFunction(p.condL)
		}

		return lua.LNil

	case "env":
		switch key {
		case "bind":
			return lua.NewFunction(p.bindL)
		case "api":
			return lua.NewFunction(p.apiL)
		case "where":
			return lua.NewFunction(p.EnvWhereL)
		}
		L.RaiseError("not found env property %s", key)
		return p

	case "lines":
		switch key {
		case "default":
			return lua.NewFunction(p.DefaultLinesL)
		}
		return p.propertyL(L, key)

	default:
		return p.propertyL(L, key)
	}
}

func ParamIndexL(L *lua.LState, key string) lua.LValue {
	switch key {
	case "checkbox":
		return &Param{Mime: "checkbox"}

	case "radio":
		return &Param{Mime: "radio"}

	case "select", "_select":
		return &Param{Mime: "select"}

	case "string":
		param := &Param{Mime: "string"}
		return param

	case "input":
		param := &Param{Mime: "input"}
		return param

	case "textarea", "text":
		param := &Param{Mime: "textarea"}
		return param

	case "lines":
		param := &Param{Mime: "lines"}
		return param

	case "lua":
		param := &Param{Mime: "lua"}
		return param

	case "int":
		param := &Param{Mime: "int"}
		return param

	case "url":
		param := &Param{Mime: "url"}
		return param

	case "network":
		param := &Param{Mime: "network"}
		return param

	case "array":
		param := &Param{Mime: "array"}
		return param
	case "map":
		param := &Param{Mime: "map"}
		return param

	case "attach":
		param := &Param{Mime: "attach"}
		return param

	case "env":
		param := &Param{Mime: "env"}
		return param
	}

	return lua.LNil
}
