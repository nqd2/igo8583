// Package igo8583 packs and unpacks ISO 8583 messages using compiled field layouts.
//
// Preset factories provide ASCII and binary layouts. FromJSON and NewBuilder
// compile custom codecs. A layout defines wire encoding, not transaction rules.
// Compiled codecs can be shared when their field packagers are immutable.
// Messages and referenced field data must not be mutated during serialization.
package igo8583
