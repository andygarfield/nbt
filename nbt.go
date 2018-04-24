package main

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	f, _ := os.Open("./bigtest.nbt")

	gr, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println(err)
	}

	if readByte(gr) != 0xA {
		fmt.Println("NBT data does not begin with root compound tag")
	}

	key := getKey(gr)
	out, err := DecodeCompound(gr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(key, out)
}

func DecodeCompound(r io.Reader) (map[string]interface{}, error) {
	compound := map[string]interface{}{}

	endFound := false
	for !endFound {
		switch readByte(r) {
		case 0x0:
			fmt.Println("END")
			endFound = true
		case 0x1:
			fmt.Println("BYTE")
			key := getKey(r)
			val := readByte(r)

			compound[key] = val
		case 0x2:
			fmt.Println("SHORT")
			key := getKey(r)
			val := readInt16(r)

			compound[key] = val
		case 0x3:
			fmt.Println("INT")
			key := getKey(r)
			val := readInt32(r)

			compound[key] = val
		case 0x4:
			fmt.Println("LONG")
			key := getKey(r)
			val := readInt64(r)

			compound[key] = val
		case 0x5:
			fmt.Println("FLOAT")
			key := getKey(r)
			val := readFloat32(r)

			compound[key] = val
		case 0x6:
			fmt.Println("DOUBLE")
			key := getKey(r)
			val := readFloat64(r)

			compound[key] = val
		case 0x7:
			fmt.Println("BYTE ARRAY")
			key := getKey(r)

			len := readInt32(r)
			b := make([]byte, len)
			for i := 0; i < int(len); i++ {
				b[i] = readByte(r)
			}
			compound[key] = b
		case 0x8:
			fmt.Println("STRING")
			key := getKey(r)

			valLen := int(readInt16(r))
			val := readString(r, valLen)

			compound[key] = val
		case 0x9:
			fmt.Println("LIST")
			key := getKey(r)
			fmt.Println("Key is:", key)
			val := DecodeList(r)

			compound[key] = val
		case 0xA:
			fmt.Println("COMPOUND")
			key := getKey(r)

			val, err := DecodeCompound(r)
			if err != nil {
				fmt.Println(err)
			}

			compound[key] = val
		case 0xB:
			fmt.Println("INT ARRAY")
			key := getKey(r)

			len := readInt32(r)
			b := make([]int32, len)
			for i := 0; i < int(len); i++ {
				b[i] = readInt32(r)
			}
			compound[key] = b
		case 0xC:
			fmt.Println("LONG ARRAY")
			key := getKey(r)

			len := readInt32(r)
			b := make([]int64, len)
			for i := 0; i < int(len); i++ {
				b[i] = readInt64(r)
			}
			compound[key] = b
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
	fmt.Println("Length is:", ln)

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
