// History: Sep 03 13 tcolar Creation

package phpserializer

import(
  "log"
  "testing"
  "strings"
  "bytes"
)

//  Test decoding a seialized PHP string into a Go structure
func TestDecoder(t *testing.T) {

  var obj = testStruct{Dummy : "Dummy"} // instance

  r := strings.NewReader(testData)

  // decode the string into target object
  p := PhpSerializer{
    // Optional, those two are the default
    // But you could implement your own if needed
    EncodeNameConverter : &SnakeToUnderscore{},
    DecodeNameConverter : &UnderscoreToSnake{},
  }
  err := p.Decode(r, &obj)

  if err != nil {
    log.Print(obj)
    panic(err)
  }
    log.Print(obj)

  // check values
  if obj.BulkLength != "8" {log.Fatal("Wrong BulkLength value: "+obj.BulkLength)}
  if obj.MinimumQty != "1" {log.Fatal("Wrong MinimumQty value: "+obj.MinimumQty)}
  if obj.MinimumQtyRestrict != 1 {log.Fatal("Wrong MinimumQtyRestrict value.")}
  if obj.Wholesale != "12" {log.Fatal("Wrong Wholesale value: "+obj.Wholesale)}
  if obj.Dummy != "Dummy" {log.Fatal("Wrong Dummy value: "+obj.Dummy)}
  if len(obj.Products) != 4 {log.Fatal("Wrong number of products.")}
  if obj.Products[419] != "419" {log.Fatal("Wrong value for product #419: " + obj.Products[419])}
  if obj.Products[420] != "420" {log.Fatal("Wrong value for product #420: " + obj.Products[420])}
  if obj.Products[421] != "421" {log.Fatal("Wrong value for product #421: " + obj.Products[421])}
  if obj.Products[1255] != "1255" {log.Fatal("Wrong value for product #1255: " + obj.Products[1255])}
}

func TestEncoder(t *testing.T) {
 // instance
  var obj = testStruct{
    Dummy : "Dummy", BulkLength : "8", MinimumQty : "1",
    MinimumQtyRestrict: 1, ApplyTo : "subtotal", Wholesale : "12",
  }
  obj.Products = map[int]string{419:"419", 420:"420", 421:"421", 1255:"1255"}

  out := []byte{}
  buffer := bytes.NewBuffer(out)

  p := PhpSerializer{}
  err := p.Encode(obj, buffer)

  if err != nil {
    panic(err)
  }

  if buffer.String() != testData2 {
    log.Print(testData2)
    log.Print(buffer.String())

    log.Fatal("Serialized data does not match expected result.")
  }
}

// test serialized string
var testData string =`a:7:{s:11:"bulk_length";s:1:"8";s:11:"minimum_qty";s:1:"1";s:20:"minimum_qty_restrict";i:1;s:17:"max_uses_per_user";s:1:"1";s:8:"apply_to";s:8:"subtotal";s:8:"products";a:4:{i:419;s:3:"419";i:420;s:3:"420";i:421;s:3:"421";i:1255;s:4:"1255";}s:9:"wholesale";s:2:"12";}`
var testData2 string =`a:7:{s:11:"bulk_length";s:1:"8";s:11:"minimum_qty";s:1:"1";s:20:"minimum_qty_restrict";i:1;s:8:"apply_to";s:8:"subtotal";s:8:"products";a:4:{i:419;s:3:"419";i:420;s:3:"420";i:421;s:3:"421";i:1255;s:4:"1255";}s:9:"wholesale";s:2:"12";s:5:"dummy";s:5:"Dummy";}`

// Test Data structure that we will Serialize / DeSerialize into
type testStruct struct { // Structure we will deserailize into
  BulkLength string
  MinimumQty string
  MinimumQtyRestrict int
  ApplyTo string
  Products map[int]string
  Wholesale string

  Dummy string
}


