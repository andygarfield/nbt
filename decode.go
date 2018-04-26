package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type Nbt struct {
	Name string
	Root map[string]struct {
		Level       interface{}
		DataVersion interface{}
	}
}

func main() {
	f, _ := os.Open("./test.nbt")

	//	gr, err := gzip.NewReader(f)
	//	if err != nil {
	//		fmt.Println(err)
	//	}

	var out Nbt

	Unmarshal(f, &out)

	//fmt.Println(out)
}

func Unmarshal(r io.Reader, v interface{}) error {
	if readByte(r) != 0xA {
		fmt.Println("NBT data does not begin with root compound tag")
	}

	destStructVal := reflect.ValueOf(v).Elem()
	destType := destStructVal.Type()

	// Check for required fields
	hasNameField := false
	hasRootField := false
	for i := 0; i < destStructVal.NumField(); i++ {
		f := destStructVal.Field(i)
		println(f.Type().String())
		if destType.Field(i).Name == "Name" && f.Type().String() == "string" {
			hasNameField = true
			key := getKey(r)
			fmt.Println("It has a name", key)
			f.SetString(key)
		}
		if destType.Field(i).Name == "Root" && f.Type().String() == "map" {
			hasRootField = true
			fmt.Println("It has a root")
		}
	}
	if !hasNameField {
		fmt.Println("Not saving the name string:", getKey(r))
	}
	if !hasRootField {
		return errors.New("Has no 'Root' field")
	}
	//compoundVal, _ := DecodeCompound(r)

	return nil
}

func DecodeCompound(r io.Reader) (map[string]interface{}, error) {
	compound := map[string]interface{}{}

	endFound := false
	for !endFound {

		typeByte := readByte(r)
		if typeByte == 0x0 {
			endFound = true
		} else {
			key := getKey(r)

			switch readByte(r) {
			case 0x1:
				val := readByte(r)
				compound[key] = val
			case 0x2:
				val := readInt16(r)
				compound[key] = val
			case 0x3:
				val := readInt32(r)
				compound[key] = val
			case 0x4:
				val := readInt64(r)
				compound[key] = val
			case 0x5:
				val := readFloat32(r)
				compound[key] = val
			case 0x6:
				val := readFloat64(r)
				compound[key] = val
			case 0x7:
				len := readInt32(r)
				val := make([]byte, len)
				for i := 0; i < int(len); i++ {
					val[i] = readByte(r)
				}

				compound[key] = val
			case 0x8:
				valLen := int(readInt16(r))
				val := readString(r, valLen)

				compound[key] = val
			case 0x9:
				val := DecodeList(r)

				compound[key] = val
			case 0xA:
				val, err := DecodeCompound(r)
				if err != nil {
					fmt.Println(err)
				}

				compound[key] = val
			case 0xB:
				len := readInt32(r)
				val := make([]int32, len)
				for i := 0; i < int(len); i++ {
					val[i] = readInt32(r)
				}
				compound[key] = val
			case 0xC:
				len := readInt32(r)
				val := make([]int64, len)
				for i := 0; i < int(len); i++ {
					val[i] = readInt64(r)
				}
				compound[key] = val
			}
		}
	}

	return compound, nil
}

func DecodeList(r io.Reader) []interface{} {
	t := readByte(r)

	ln := int(readInt32(r))
	if ln <= 0 {
		return []interface{}{}
	}

	list := make([]interface{}, ln)

	switch t {
	case 0x0:
		return []interface{}{}
	case 0x1:
		for i := 0; i < ln; i++ {
			list[i] = readByte(r)
		}
	case 0x2:
		for i := 0; i < ln; i++ {
			list[i] = readInt16(r)
		}
	case 0x3:
		for i := 0; i < ln; i++ {
			list[i] = readInt32(r)
		}
	case 0x4:
		for i := 0; i < ln; i++ {
			list[i] = readInt64(r)
		}
	case 0x5:
		for i := 0; i < ln; i++ {
			list[i] = readFloat32(r)
		}
	case 0x6:
		for i := 0; i < ln; i++ {
			list[i] = readFloat64(r)
		}
	case 0x7:
		for i := 0; i < ln; i++ {
			aLn := int(readInt32(r))
			nl := make([]byte, aLn)
			for j := 0; j < aLn; j++ {
				nl[j] = readByte(r)
			}
			list[i] = nl
		}
	case 0x8:
		for i := 0; i < ln; i++ {
			sLn := int(readInt16(r))
			list[i] = readString(r, sLn)
		}
	case 0x9:
		for i := 0; i < ln; i++ {
			aLn := int(readInt32(r))
			nl := make([]interface{}, aLn)
			for j := 0; j < aLn; j++ {
				nl[j] = DecodeList(r)
			}
		}
	case 0xA:
		for i := 0; i < ln; i++ {
			c, err := DecodeCompound(r)
			if err != nil {
				fmt.Println(err)
			}
			list[i] = c
		}
	case 0xB:
		for i := 0; i < ln; i++ {
			aLn := int(readInt32(r))
			nl := make([]int32, aLn)
			for j := 0; j < aLn; j++ {
				nl[j] = readInt32(r)
			}
			list[i] = nl
		}
	case 0xC:
		for i := 0; i < ln; i++ {
			aLn := int(readInt32(r))
			nl := make([]int64, aLn)
			for j := 0; j < aLn; j++ {
				nl[j] = readInt64(r)
			}
			list[i] = nl
		}
	}

	return list
}

func getKey(r io.Reader) string {
	keyLen := int(readInt16(r))
	return readString(r, keyLen)
}

func readByte(r io.Reader) byte {
	b := make([]byte, 1)
	r.Read(b)
	return b[0]
}

func readString(r io.Reader, l int) string {
	b := make([]byte, l)
	r.Read(b)
	return strings.TrimSpace(string(b))
}

func readInt8(r io.Reader) int8 {
	b := make([]byte, 1)
	r.Read(b)
	return int8(b[0])
}

func readInt16(r io.Reader) int16 {
	var s int16
	binary.Read(r, binary.BigEndian, &s)
	return s
}

func readInt32(r io.Reader) int32 {
	var s int32
	binary.Read(r, binary.BigEndian, &s)
	return s
}

func readInt64(r io.Reader) int64 {
	var s int64
	binary.Read(r, binary.BigEndian, &s)
	return s
}

func readFloat32(r io.Reader) float32 {
	var f float32
	binary.Read(r, binary.BigEndian, &f)
	return f
}

func readFloat64(r io.Reader) float64 {
	var f float64
	binary.Read(r, binary.BigEndian, &f)
	return f
}

func skipBytes(r io.Reader, l int) {
	b := make([]byte, l)
	r.Read(b)
}
