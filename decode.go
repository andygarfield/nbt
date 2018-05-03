package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
)

const (
	endTag = iota
	byteTag
	shortTag
	intTag
	longTag
	floatTag
	doubleTag
	byteArrayTag
	stringTag
	listTag
	compoundTag
	intArrayTag
	longArrayTag
)

type HelloWorld struct {
	Sample    string
	OtherVals struct {
		Rando     string
		FavNumber int64
	}
}

func main() {
	f, _ := os.Open("./byte.nbt")

	//	gr, err := gzip.NewReader(f)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	v := HelloWorld{}
	Unmarshal(f, &v)
	fmt.Println(v)
}

// Unmarshal decodes NBT data coming from stream `r` and decodes it into
// destination `v`, which must be a pointer to a `struct` or `interface{}`.
// If `v` is of type struct, the struct must have the same name of the root
// compound tag, in Pascal case (i.e. `BeginWithUpperCase`).
func Unmarshal(r io.Reader, v interface{}) error {
	if readByte(r) != 0xA {
		fmt.Println("NBT data does not begin with root compound tag")
	}

	structVal := reflect.ValueOf(v)

	compoundName := toPascalCase(getKey(r))
	structName := getStructName(structVal)

	fmt.Printf("Compound Key: %v\nStructKey: %v\n", compoundName, structName)
	if structName != compoundName {
		return fmt.Errorf("Struct name (%v) does not match compound name (%v)", structName, toPascalCase(compoundName))
	}

	unmarshalCompound(r, structVal)

	return nil
}

// unmarshalCompound takes a io.Reader that has initiated a Compound tag and modifies v.
// For each named key in the Compound tag, the fields of v is checked. If there is no
// match, the loop continues. If there is a match, the Compound tag type is compared
// against the type in v. If the type in v is not appropriate, given the value, the
// function returns an error. If the type in v is an `interface{}`, then the value is
// stored directly.
func unmarshalCompound(r io.Reader, structVal reflect.Value) error {
	for {
		compoundType := readByte(r)
		fmt.Println("Compound Type:", compoundType)
		if compoundType == endTag {
			break
		}
		matchedField := getMatchingField(r, structVal)

		switch compoundType {
		case byteTag:
			matchedField.SetUint(uint64(readByte(r)))
		case shortTag:
			newShort := int64(readInt16(r))
			fmt.Println(newShort)
			matchedField.SetInt(newShort)
		case intTag:
			matchedField.SetInt(int64(readInt32(r)))
		case longTag:
			matchedField.SetInt(int64(readInt64(r)))
		case floatTag:
			matchedField.SetFloat(float64(readFloat32(r)))
		case doubleTag:
			matchedField.SetFloat(readFloat64(r))
		case byteArrayTag:
			len := readInt32(r)
			val := make([]byte, len)
			for i := 0; i < int(len); i++ {
				val[i] = readByte(r)
			}

			matchedField.SetBytes(val)
		case stringTag:
			valLen := int(readInt16(r))
			matchedField.SetString(readString(r, valLen))
			//			case listTag:
			//				compound[key] = DecodeList(r)
		case compoundTag:
			err := unmarshalCompound(r, matchedField)
			fmt.Printf("Struct name: %v\n", matchedField)
			if err != nil {
				fmt.Println(err)
			}
			//
			//				compound[key] = val
			//			case intArrayTag:
			//				len := readInt32(r)
			//				val := make([]int32, len)
			//				for i := 0; i < int(len); i++ {
			//					val[i] = readInt32(r)
			//				}
			//				compound[key] = val
			//			case longArrayTag:
			//				len := readInt32(r)
			//				val := make([]int64, len)
			//				for i := 0; i < int(len); i++ {
			//					val[i] = readInt64(r)
			//				}
			//				compound[key] = val
		}
	}

	return nil
}

//func DecodeList(r io.Reader) []interface{} {
//	t := readByte(r)
//
//	ln := int(readInt32(r))
//	if ln <= 0 {
//		return []interface{}{}
//	}
//
//	list := make([]interface{}, ln)
//
//	switch t {
//	case 0x0:
//		return []interface{}{}
//	case 0x1:
//		for i := 0; i < ln; i++ {
//			list[i] = readByte(r)
//		}
//	case 0x2:
//		for i := 0; i < ln; i++ {
//			list[i] = readInt16(r)
//		}
//	case 0x3:
//		for i := 0; i < ln; i++ {
//			list[i] = readInt32(r)
//		}
//	case 0x4:
//		for i := 0; i < ln; i++ {
//			list[i] = readInt64(r)
//		}
//	case 0x5:
//		for i := 0; i < ln; i++ {
//			list[i] = readFloat32(r)
//		}
//	case 0x6:
//		for i := 0; i < ln; i++ {
//			list[i] = readFloat64(r)
//		}
//	case 0x7:
//		for i := 0; i < ln; i++ {
//			aLn := int(readInt32(r))
//			nl := make([]byte, aLn)
//			for j := 0; j < aLn; j++ {
//				nl[j] = readByte(r)
//			}
//			list[i] = nl
//		}
//	case 0x8:
//		for i := 0; i < ln; i++ {
//			sLn := int(readInt16(r))
//			list[i] = readString(r, sLn)
//		}
//	case 0x9:
//		for i := 0; i < ln; i++ {
//			aLn := int(readInt32(r))
//			nl := make([]interface{}, aLn)
//			for j := 0; j < aLn; j++ {
//				nl[j] = DecodeList(r)
//			}
//		}
//	case 0xA:
//		for i := 0; i < ln; i++ {
//			c, err := unmarshalCompound(r)
//			if err != nil {
//				fmt.Println(err)
//			}
//			list[i] = c
//		}
//	case 0xB:
//		for i := 0; i < ln; i++ {
//			aLn := int(readInt32(r))
//			nl := make([]int32, aLn)
//			for j := 0; j < aLn; j++ {
//				nl[j] = readInt32(r)
//			}
//			list[i] = nl
//		}
//	case 0xC:
//		for i := 0; i < ln; i++ {
//			aLn := int(readInt32(r))
//			nl := make([]int64, aLn)
//			for j := 0; j < aLn; j++ {
//				nl[j] = readInt64(r)
//			}
//			list[i] = nl
//		}
//	}
//
//	return list
//}

func getMatchingField(r io.Reader, structVal reflect.Value) reflect.Value {
	fieldKey := toPascalCase(getKey(r))
	fmt.Println("Field key is:", fieldKey)

	matchedField := reflect.Indirect(structVal).FieldByName(fieldKey)

	if matchedField.Kind() == reflect.Invalid {
		fmt.Println("Invalid field")
	}

	return matchedField
}

func toPascalCase(raw string) string {
	reg := regexp.MustCompile("[a-zA-Z]+")
	title := strings.Title(raw)
	ss := strings.Split(title, " ")

	out := ""
	for _, s := range ss {
		rs := []rune(s)
		for _, r := range rs {
			if reg.MatchString(string(r)) {
				out += string(r)
			}
		}
	}
	return out
}

func getStructName(structVal reflect.Value) string {
	structType := reflect.Indirect(structVal).Type()
	structName := structType.Name()

	return structName
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
