// History: Sep 03 13 tcolar Creation

package phpserializer

import(
  "log"
  "testing"
  "strings"
)

//  Test decoding a seialized PHP string into a Go structure
func TestDecoder(t *testing.T) {

  type target struct { // Structure we will deserailize into
    BulkLength string
    MinimumQty string
    MinimumQtyRestrict int
    ApplyTo string
    Wholesale string
    Products map[int]string

    Dummy string
  }
  var obj = target{Dummy : "Dummy"} // instance

  str :=`a:7:{s:11:"bulk_length";s:1:"8";s:11:"minimum_qty";s:1:"1";s:20:"minimum_qty_restrict";i:1;s:17:"max_uses_per_user";s:1:"1";s:8:"apply_to";s:8:"subtotal";s:8:"products";a:4:{i:419;s:3:"419";i:420;s:3:"420";i:421;s:3:"421";i:1255;s:4:"1255";}s:9:"wholesale";s:2:"12";}`
  r := strings.NewReader(str)

  // decode the string into target object
  p := PhpSerializer{}
  err := p.Decode(r, &obj)

  if err != nil {
    log.Print(obj)
    panic(err)
  }

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

