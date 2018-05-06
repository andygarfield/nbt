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
	StSeq []string
}

func main() {
	f, _ := os.Open("./byte.nbt")

	//	gr, err := gzip.NewReader(f)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	v := HelloWorld{}
	Unmarshal(f, &v)
	fmt.Printf("%#v\n", v)
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
		fmt.Println("Matched field:", matchedField)
		decodeValue(r, matchedField, compoundType)
	}

	return nil
}

func decodeValue(r io.Reader, val reflect.Value, tagType byte) {
	switch tagType {
	//	case endTag:
	//		return
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
