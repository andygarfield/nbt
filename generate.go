package main

import (
	"fmt"
	"io"
	"reflect"
)

func decodeValue(r io.Reader, tagType byte) {
	switch tagType {
	case endTag:
		return
	case byteTag:
		v := uint64(readByte(r))
		val.SetUint(v)
	case shortTag:
		newShort := int64(readInt16(r))
		val.SetInt(newShort)
	case intTag:
		val.SetInt(int64(readInt32(r)))
	case longTag:
		val.SetInt(int64(readInt64(r)))
	case floatTag:
		val.SetFloat(float64(readFloat32(r)))
	case doubleTag:
		val.SetFloat(readFloat64(r))
	case byteArrayTag:
		t := val.Type()
		ln := int(readInt32(r))
		s := reflect.MakeSlice(t, ln, ln)
		for i := 0; i < int(ln); i++ {
			decodeValue(r, s.Index(i), byteTag)
		}
		val.Set(s)
	case stringTag:
		valLen := int(readInt16(r))
		val.SetString(readString(r, valLen))
	case listTag:
		listType := readByte(r)
		t := val.Type()
		ln := int(readInt32(r))
		s := reflect.MakeSlice(t, ln, ln)
		for i := 0; i < int(ln); i++ {
			decodeValue(r, s.Index(i), listType)
		}
		val.Set(s)
	case compoundTag:
		err := unmarshalCompound(r, val)
		fmt.Printf("Struct name: %v\n", val)
		if err != nil {
			fmt.Println(err)
		}
	case intArrayTag:
		t := val.Type()
		ln := int(readInt32(r))
		s := reflect.MakeSlice(t, ln, ln)
		for i := 0; i < int(ln); i++ {
			decodeValue(r, s.Index(i), intTag)
		}
		val.Set(s)
	case longArrayTag:
		t := val.Type()
		ln := int(readInt32(r))
		s := reflect.MakeSlice(t, ln, ln)
		for i := 0; i < int(ln); i++ {
			decodeValue(r, s.Index(i), longTag)
		}
		val.Set(s)
	}
}
