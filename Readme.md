PHPSerializer for Go
======

Purpose
===

This package provide some Features to decode/encode PHP serialialized objects into
Go structures.

See [phpserializer_test.go](phpserializer_test.go) for usage examples.

But Why ?
===

At work (Rivet & Sway) I needed to handle some data from a Drupal database from Go.

Among all the terrible things Drupal does one particular painful one is that it stores
serialized PHP objects in the database (in text fields) ... Yeah, really !

Added to that there is apparently **NO** official PHP serializtion format specifications,
other than checking the PHP source code and crying in the process.

So anyway this allows Encoding/Decoding Serialized PHP object strings into Go structures.
There is no guarantees it supports all possible PHP objects at this point, pull requests welcome.

