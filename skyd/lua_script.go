package skyd

/*
#cgo LDFLAGS: -lluajit-5.1.2
#include <stdlib.h>
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
*/
import "C"

import (
  "bytes"
  "errors"
  "fmt"
  "regexp"
  "sort"
  "text/template"
  "unsafe"
)

// A LuaScript is used to bridge between Go and LuaJIT.
type LuaScript struct {
  state         *C.lua_State
  header        string
  source        string
  propertyFile  *PropertyFile
}

func NewLuaScript(propertyFile *PropertyFile, source string) *LuaScript {
  if propertyFile == nil {
    panic("skyd.LuaScript: Property file required.")
  }
  return &LuaScript{propertyFile:propertyFile, source:source}
}

// Retrieves the source for the script.
func (l *LuaScript) Source() string {
  return l.source
}

// Retrieves the generated header for the script.
func (l *LuaScript) Header() string {
  return l.header
}

// Executes the lua script and returns the data.
func (l *LuaScript) Execute(functionName string, nresult int, v ...interface{}) ([]interface{}, error) {
  if err := l.Init(); err != nil {
    return nil, err
  }

  // Execute function and encode the return value as .
  return nil, nil
}

// Initializes the Lua context and compiles the source code.
func (l *LuaScript) Init() error {
  if l.state != nil {
    return nil
  }

  // Initialize the state and open the libraries.
  l.state = C.luaL_newstate()
  if l.state == nil {
    return errors.New("Unable to initialize Lua context.")
  }
  C.luaL_openlibs(l.state);
  
  // Generate the header file.
  err := l.generateHeader()
  if err != nil {
    l.Destroy()
    return err
  }

  // Compile the script.
  source := C.CString(l.source)
  defer C.free(unsafe.Pointer(source))
  ret := C.luaL_loadstring(l.state, source)
  if ret != 0 {
    defer l.Destroy()
    errstring := C.GoString(C.lua_tolstring(l.state, -1, nil))
    return fmt.Errorf("skyd.LuaScript: Syntax Error: %v", errstring)
  }

  // Run script once to initialize.
  ret = C.lua_pcall(l.state, 0, 0, 0);
  if ret != 0 {
    defer l.Destroy()
    errstring := C.GoString(C.lua_tolstring(l.state, -1, nil))
    return fmt.Errorf("skyd.LuaScript: Init Error: %v", errstring)
  }
  
  return nil
}

// Closes the lua context.
func (l *LuaScript) Destroy() {
  if l.state != nil {
    C.lua_close(l.state)
    l.state = nil
  }
}

// Generates the header for the script based on a source string.
func (l *LuaScript) generateHeader() error {
  // Find a list of all references properties.
  properties, err := l.ExtractPropertyReferences()
  if err != nil {
    return err
  }
  
  // Parse the header template.
  t := template.New("header.lua")
  t.Funcs(template.FuncMap{"structdef": propertyStructDef, "metatypedef": metatypeFunctionDef, "initdescriptor": initDescriptorDef})
  t, err = t.ParseFiles("tmpl/header.lua")
  if err != nil {
    return err
  }
  
  // Generate the template from the property references.
  var buffer bytes.Buffer
  err = t.Execute(&buffer, properties)
  if err != nil {
    return err
  }
  
  // Assign header
  l.header = buffer.String()
  
  return nil
}

// Extracts the property references from the source string.
func (l *LuaScript) ExtractPropertyReferences() ([]*Property, error) {
  // Create a list of properties.
  properties := make([]*Property, 0)
  lookup := make(map[int64]*Property)

  // Find all the event property references in the script.
  r, err := regexp.Compile(`\bevent(?:\.|:)(\w+)`)
  if err != nil {
    return nil, err
  }
  for _, match := range r.FindAllStringSubmatch(l.source, -1) {
    name := match[1]
    property := l.propertyFile.GetPropertyByName(name)
    if property == nil {
      return nil, fmt.Errorf("Property not found: '%v'", name)
    }
    if lookup[property.Id] == nil {
      properties = append(properties, property)
      lookup[property.Id] = property
    }
  }
  sort.Sort(PropertyList(properties))

  return properties, nil
}


func propertyStructDef(args ...interface{}) string {
  if property, ok := args[0].(*Property); ok {
    switch property.DataType {
    case StringDataType: return fmt.Sprintf("sky_string_t _%v;", property.Name)
    case IntegerDataType: return fmt.Sprintf("int32_t _%v;", property.Name)
    case FloatDataType: return fmt.Sprintf("double _%v;", property.Name)
    case BooleanDataType: return fmt.Sprintf("bool _%v;", property.Name)
    }
  }
  return ""
}

func metatypeFunctionDef(args ...interface{}) string {
  if property, ok := args[0].(*Property); ok {
    switch property.DataType {
    case StringDataType: return fmt.Sprintf("%v = function(event) return ffi.string(event._%v.data, event._%v.length) end,", property.Name, property.Name, property.Name)
    default: return fmt.Sprintf("%v = function(event) return event._%v end,", property.Name, property.Name)
    }
  }
  return ""
}

func initDescriptorDef(args ...interface{}) string {
  if property, ok := args[0].(*Property); ok {
    return fmt.Sprintf("descriptor:set_property(%d, ffi.offsetof('sky_lua_event_t', '_%s'), '%s')", property.Id, property.Name, property.DataType)
  }
  return ""
}