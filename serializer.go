// History: Sep 03 13 Thibaut Colar Creation

package phpserializer

import(
  "log"
  "io"
  "fmt"
  "errors"
  "reflect"
  "strconv"
  "text/scanner"
)

// Encode / Decode serialized PHP objects from Go / Golang
// Note that there are apparently **NO** PHP serialization format specs, so this might not be perfect.
// Defaults to SnakeToUnderscore & UnderscoreToSnake for Name conversion, but can be replaced.
type PhpSerializer struct{
  EncodeNameConverter NameConverter
  DecodeNameConverter NameConverter
}

// Decode serialized PHP content into target object
// Note: Not closing the reader
func (p PhpSerializer) Decode(reader io.Reader, v interface{}) (err error) {
  if p.DecodeNameConverter == nil{
    p.DecodeNameConverter = &UnderscoreToSnake{}
  }
  s := &scanner.Scanner{}
  s.Init(reader)
  obj := reflect.ValueOf(v)
  err = p.decode(s, obj, false)
  return err
}

func (p PhpSerializer) Encode(v interface{}, writer io.Writer) (err error) {
  if p.EncodeNameConverter == nil{
    p.EncodeNameConverter = &SnakeToUnderscore{}
  }
  return p.encode(v, writer)
}

// #################### Internals ############################################

// Encode a Go object into a serialized PHP and write it to the writer
// Note: Not closing the writer
func (p PhpSerializer) encode(v interface{}, writer io.Writer) (err error) {
  if p.EncodeNameConverter == nil{
    p.EncodeNameConverter = &SnakeToUnderscore{}
  }
  t := reflect.TypeOf(v)
  switch t.Kind(){
    case reflect.Struct:
      val := reflect.ValueOf(v)
      err = p.encodeStruct(val , writer)
    case reflect.Int: // TODO: are int8, int16 etc.. separate ?
      i, _ := v.(int)
      err = p.encodeInt(i , writer)
    case reflect.String:
      str, _ := v.(string)
      err = p.encodeString(str , writer)
    case reflect.Map:
      val := reflect.ValueOf(v)
      err = p.encodeMap(val , writer)
    default:
      return errors.New(fmt.Sprintf("Unsupported element %s - of type %s",
                  reflect.ValueOf(v).String(),
                  t.String()))
  }
  return err
}

// Encode a structure
func (p PhpSerializer) encodeStruct(v reflect.Value, writer io.Writer) (err error) {
  size := v.NumField()
  p.write(writer, fmt.Sprintf("a:%d:{", size))
  for i := 0; i < size; i++ {
    // field name
    fieldName := p.EncodeNameConverter.Convert(v.Type().Field(i).Name)
    err = p.encodeString(fieldName, writer)
    if err != nil {return err}
    // field value
    field := v.Field(i)
    err = p.Encode(field.Interface(), writer)
    if err != nil {return err}
  }
  err = p.write(writer, "}")
  return err
}

// Encode a string
func (p PhpSerializer) encodeString(str string, writer io.Writer) (err error) {
  err = p.write(writer, fmt.Sprintf(`s:%d:"%s";`, len(str), str))
  return err
}

// Encode an int
func (p PhpSerializer) encodeInt(i int, writer io.Writer) (err error) {
  err = p.write(writer, fmt.Sprintf("i:%d;", i))
  return err
}

// Encode a map
func (p PhpSerializer) encodeMap(v reflect.Value, writer io.Writer) (err error) {
  size := len(v.MapKeys())
  p.write(writer, fmt.Sprintf("a:%d:{", size))
  for _, key := range v.MapKeys(){
    err = p.Encode(key.Interface(), writer);
    if err != nil {return err}
    val := v.MapIndex(key)
    err = p.Encode(val.Interface(), writer);
    if err != nil {return err}
  }
  err = p.write(writer, "}")
  return err
}

// write a string to the writer
func (p PhpSerializer) write(writer io.Writer, str string) (err error) {
  _, err = writer.Write([]byte(str))
  return err
}

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
    case 'd': // decimal
      err = p.decodeFloat(s, v, skipVal)
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
  if err = p.decodeToken(s, "a"); err != nil {return err}
  size, err = p.decodeSize(s)
  if err != nil {return err}
  if err = p.decodeToken(s, "{"); err != nil {return err}
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
        // Create an instance of the map value type
        mapVal := reflect.New(v.Type().Elem()).Elem()
        // and then decode that
        err = p.decode(s, mapVal, skipVal)
        if err != nil {return err}
        v.SetMapIndex(reflect.ValueOf(key), mapVal)
      } else {
        // targetting a struct
        var k string
        k, _ = key.(string)
        if k == "" {return errors.New(fmt.Sprintf("Struct Key '%s' is not a string.", key))}
        target := p.DecodeNameConverter.Convert(k)
        var val reflect.Value
        if v.Kind() == reflect.Struct {
          val = v.FieldByName(target)
        } else {
          val = v.Elem().FieldByName(target)
        }
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
  if err = p.decodeToken(s, "i"); err != nil {return err}
  if err = p.decodeToken(s, ":"); err != nil {return err}
  s.Scan(); text := s.TokenText()
  var i int
  nb, err := strconv.Atoi(text)
  if err != nil {return err}
  i = nb
  if err = p.decodeToken(s, ";"); err != nil {return err}
  if ! skipVal {v.Set(reflect.ValueOf(i))}
  return err
}

// Decode a float
func (p PhpSerializer) decodeFloat(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error){
  if err = p.decodeToken(s, "d"); err != nil {return err}
  if err = p.decodeToken(s, ":") ; err != nil {return err}
  s.Scan(); text := s.TokenText()
  var f float64
  nb, err := strconv.ParseFloat(text, 64)
  if err != nil {return err}
  f = nb
  if err = p.decodeToken(s, ";"); err != nil {return err}
  if ! skipVal {v.Set(reflect.ValueOf(f))}
  return err
}

// Decode a string into v
func (p PhpSerializer) decodeString(s *scanner.Scanner, v reflect.Value, skipVal bool) (err error){
  if err = p.decodeToken(s, "s"); err != nil {return err}
  size, err := p.decodeSize(s)
  if err != nil {return err}
  str := "";
  for len(str) < size + 2{
    str += string(s.Next())
  }
  if str[0] != '"' || str[len(str) - 1] != '"'{
    return p.error(s, "Expected string to be surrounded by quotes !", str)
  }
  if err = p.decodeToken(s, ";"); err != nil {return err}
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
  if err = p.decodeToken(s, ":"); err != nil {return size, err}
  s.Scan(); val := s.TokenText()
  size, err = strconv.Atoi(val)
  if err != nil {return size, p.error(s, err.Error(), val)}
  if err = p.decodeToken(s, ":"); err != nil {return size, err}
  return size, err
}

// Create a scanning error
func (p PhpSerializer) error(s *scanner.Scanner, msg string, lastTokenText string) error{
  pos := s.Pos()
  msg += fmt.Sprintf(". At Position(%d,%d), got '%s'", pos.Line, pos.Column, lastTokenText)
  return errors.New(msg)
}


