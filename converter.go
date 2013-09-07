// History: Sep 06 13 tcolar Creation

package phpserializer

import(
  "strings"
  "unicode"
)

// Converter interface, takes a string and convert to another
type NameConverter interface {
  Convert(str string) string
}

// Convert an underscored name to a snake case name
// ie: "this_is_an_example" -> "ThisIsAnExample"
type UnderscoreToSnake struct{}

func (c UnderscoreToSnake) Convert(str string) string {
  snaked := []string{""}
  for _, part := range strings.Split(str, "_") {
    snaked = append(snaked, strings.Title(part))
  }
  return strings.Join(snaked, "")
}

// Convert snake case name to a lower case underscored name
// ie: "ThisIsAnExample" -> "this_is_an_example"
type SnakeToUnderscore struct{}

func (c SnakeToUnderscore) Convert(str string) string {
  result := []rune{}
  for _, c := range str {
    if len(result) > 0 && unicode.IsUpper(c){
      result = append(result, '_')
    }
    result = append(result, unicode.ToLower(c))
  }
  return string(result)
}

