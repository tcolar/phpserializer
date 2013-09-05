// History: Sep 03 13 Thibaut Colar Creation

package phpserializer

import(
  "log"
  "io"
  "fmt"
  "errors"
  "reflect"
  "strconv"
  "strings"
  "text/scanner"
)

// Encode / Decode serialized PHP objects from Go / Golang
// Note that there are apparently **NO** PHP serialization format specs, so this might not be perfect.
type PhpSerializer struct{
}

// Decode serialized PHP content into target object
func (p PhpSerializer) Decode(reader io.Reader, v interface{}) (err error) {
  // http://stackoverflow.com/questions/6395076/in-golang-using-reflect-how-do-you-set-the-value-of-a-struct-field
  s := &scanner.Scanner{}
  s.Init(reader)
  obj := reflect.ValueOf(v)
  err = p.decode(s, obj, false)
  return err
}

// #################### Internals

// Decode the next element found by the scanner and set value into v
func (p PhpSerializer) decode(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error) {
  c := s.Peek()
  switch(c) {
    case 'i': // int
      err = p.decodeInt(s, v, skipVal)
    case 's': // string
      err = p.decodeString(s, v, skipVal)
    case 'a': // map / struct  TODO: can be an array/list too ?
      err = p.decodeMap(s, v, skipVal)
    //case "d": // decimal TODO: 'd' object types (float)
    //case "b": // TODO: 'b' object types (bool ??)
    //case "O": // TODO: 'O' object types
    //case "N": // TODO: 'N' object types
    default:
      s.Scan()
      err = p.error(s, "Unexpected type %s", s.TokenText())
  }
  return err
}

// Decode a map into v (map or struct)
func (p PhpSerializer) decodeMap(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error){
  var size int
  if p.decodeToken(s, "a") != nil {return err}
  size, err = p.decodeSize(s)
  if err != nil {return err}
  if p.decodeToken(s, "{") != nil {return err}
  // decode key/value pairs
  for i := 0; i != size; i++ {
    var key interface{}
    err = p.decode(s, reflect.ValueOf(&key).Elem(), skipVal)
    if err != nil {return err}
    if skipVal {
      // this is just data we need to parse but not do anything with
        err = p.skipValue(s)
        if err != nil {return err}
    } else {
      if v.Kind() == reflect.Map{
        // targetting a map
        if v.IsNil() {
          // Initialize to an empty map if nil
          v.Set(reflect.MakeMap(v.Type()))
        }
        var mapVal interface{}
        err = p.decode(s, reflect.ValueOf(&mapVal).Elem(), skipVal)
        if err != nil {return err}
        v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(mapVal))
      } else {
        // targetting a struct
        var k string
        k, _ = key.(string)
        if k == "" {return errors.New("Struct Key name is not a string.")}
        target := underscoredToSnake(k)
        val := v.Elem().FieldByName(target)
        if val.IsValid() {
          err = p.decode(s, val, skipVal)
        } else {
          log.Printf("No targets field found named '%s', skipping it !", target)
          err = p.skipValue(s)
        }
        if err != nil {return err}
      }
    }
  }
  if p.decodeToken(s, "}") != nil {return err}
  return err
}

// Read but discard a value
func (p PhpSerializer) skipValue(s *scanner.Scanner) error {
  var dummy interface{}
  return p.decode(s, reflect.ValueOf(&dummy).Elem(), true)
}

// Decode an int int v
func (p PhpSerializer) decodeInt(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error){
  if p.decodeToken(s, "i") != nil {return err}
  if p.decodeToken(s, ":") != nil {return err}
  s.Scan(); text := s.TokenText()
  var i int
  i, err = strconv.Atoi(text)
  if err != nil {return err}
  if p.decodeToken(s, ";") != nil {return err}
  if ! skipVal {v.Set(reflect.ValueOf(i))}
  return err
}

// Decode a string into v
func (p PhpSerializer) decodeString(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error){
  if p.decodeToken(s, "s") != nil {return err}
  _, err = p.decodeSize(s)
  if err != nil {return err}
  s.Scan(); str := s.TokenText()
  if str[0] != '"' || str[len(str) - 1] != '"'{
    return p.error(s, "Expected string to be surrounded by quotes !", str)
  }
  if p.decodeToken(s, ";") != nil {return err}
  str = str[1:len(str)-1]
  if ! skipVal {v.Set(reflect.ValueOf(str))}
  return err
}

// Decode a token and check against expected value
func (p PhpSerializer) decodeToken(s *scanner.Scanner, expected string) (err error){
  s.Scan(); text := s.TokenText()
  if text != expected {err = p.error(s, fmt.Sprintf("Expected '%s' !", expected), text)}
  return err
}

// Decode and return a size value (ex: ":12:")
func (p PhpSerializer) decodeSize(s *scanner.Scanner) (size int, err error){
  if p.decodeToken(s, ":") != nil {return size, err}
  s.Scan(); val := s.TokenText()
  size, err = strconv.Atoi(val)
  if err != nil {return size, p.error(s, err.Error(), val)}
  if p.decodeToken(s, ":") != nil {return size, err}
  return size, err
}

// Create a scanning error
func (p PhpSerializer) error(s *scanner.Scanner, msg string, lastTokenText string) error{
  pos := s.Pos()
  msg += fmt.Sprintf(". At Position(%d,%d), got '%s'", pos.Line, pos.Column, lastTokenText)
  return errors.New(msg)
}

// Convert an underscored name to a snake case name
// ie: "this_is_an_example" -> "ThisIsAnExample"
func underscoredToSnake(str string) string {
  snaked := []string{""}
  for _, part := range strings.Split(str, "_") {
    snaked = append(snaked, strings.Title(part))
  }
  return strings.Join(snaked, "")
}

