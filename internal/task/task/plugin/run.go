/*
 *  Copyright (c) 2022 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2022-05-19
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

/*
-- local hp, err = _M:new(chunk_size?, sock?)
function _M.new(self, chunk_size, sock)
    local state = STATE_NOT_READY

    if not sock then
        local s, err = tcp()
        if not s then
            return nil, err
        end
        sock = s
    end

    return setmetatable({
        sock = sock,
        chunk_size = chunk_size or 8192,
        total_size = 0,
        state = state,
        chunked = false,
        keepalive = true,
        _eof = false,
        previous = {},
    }, mt)
end
*/
/*
t.AddStep(&step.InstallFile{ // install tools.conf
Content:           &toolsConf,
ContainerId:       &containerName,
ContainerDestPath: "/etc/curve/tools.conf",
ExecOptions:       curveadm.ExecOptions(),
})
*/

const source = `
local task = require "task"
local step = require "step"
local options = require "options"

local host = options.get_string("host")
local mds_listen_addr = options.get_string("mds_listen_addr")

local script = "
"

local t = task.new("name", "subname", "host")

local callback = function()
end

local out = { _val = "" }
local success = { _val = false }
t:add_step(step.shell{
  command = "ls -la",
  out = out,
})
t:add_step(step.install_file{
    content = script,
})
t:add_step(step.lambda{
    lambda = callback,
})
`

type Person struct {
	Name string
}

const luaPersonTypeName = "person"

// Registers my person type to given L.
func registerPersonType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPersonTypeName)
	L.SetGlobal("person", mt)
	// static attributes
	L.SetField(mt, "new", L.NewFunction(newPerson))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), personMethods))
}

// Constructor
func newPerson(L *lua.LState) int {
	person := &Person{L.CheckString(1)}
	ud := L.NewUserData()
	ud.Value = person
	L.SetMetatable(ud, L.GetTypeMetatable(luaPersonTypeName))
	L.Push(ud)
	return 1
}

// Checks whether the first lua argument is a *LUserData with *Person and returns this *Person.
func checkPerson(L *lua.LState) *Person {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Person); ok {
		return v
	}
	L.ArgError(1, "person expected")
	return nil
}

var personMethods = map[string]lua.LGFunction{
	"name": personGetSetName,
}

// Getter and setter for the Person#Name
func personGetSetName(L *lua.LState) int {
	p := checkPerson(L)
	if L.GetTop() == 2 {
		p.Name = L.CheckString(2) // 将 lua 中的变量转换成 go 中的变量
		return 0
	}
	L.Push(lua.LString(p.Name))
	return 1
}

//////

func main() {
	L := lua.NewState()
	defer L.Close()
	registerPersonType(L)
	if err := L.DoString(`
p = person.new("Steeve")
print(p:name()) -- "Steeve"
p:name("Alice")
print(p:name()) -- "Alice"
`); err != nil {
		panic(err)
	}
}

func main() {
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("gomodule", load)
	if err := L.DoString(source); err != nil {
		panic(err)
	}

	if err := L.DoFile("test.lua"); err != nil {
		panic(err)
	}
}

func load(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.SetField(mod, "name", lua.LString("gomodule"))
	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"goFunc": goFunc,
}

func goFunc(L *lua.LState) int {
	fmt.Println("golang")
	return 0
}

/*
func Double(L *lua.LState) int {
    lv := L.ToInt(1)
    L.Push(lua.LNumber(lv * 2))
    return 1
}
*/
