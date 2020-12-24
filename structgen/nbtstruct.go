package structgen

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

// CreatePackage creates a directory and package from the nbt io.ReadSeeker input
func CreatePackage(r io.ReadSeeker, packagePath string) {
	os.Mkdir(packagePath, 0777)
	structFilename := "structs.go"
	structPath := filepath.Join(packagePath, structFilename)
	structFile, err := os.Create(structPath)
	if err != nil {
		log.Println(err)
	}

	structFile.Write([]byte(fmt.Sprintf("package %s\n", packagePath)))
	structFile.Write([]byte(createStruct(r)))

	structFile.Close()
	c := exec.Command("go", "fmt", structPath)
	if err := c.Run(); err != nil {
		log.Println(err)
	}
}

// createStruct creates the main body of the struct after decoding the nbt file
func createStruct(r io.ReadSeeker) string {
	tagType := readByte(r)
	name := toPascalCase(getKey(r))
	typeStr := getFieldType(r, tagType)

	return fmt.Sprintf("type %s %s", name, typeStr)
}

func getFieldType(r io.ReadSeeker, tagType byte) string {
	var typeStr string
	switch tagType {

	case byteTag:
		r.Seek(1, io.SeekCurrent)
		typeStr = "byte"

	case shortTag:
		r.Seek(2, io.SeekCurrent)
		typeStr = "int16"

	case intTag:
		r.Seek(4, io.SeekCurrent)
		typeStr = "int32"

	case longTag:
		r.Seek(8, io.SeekCurrent)
		typeStr = "int64"

	case floatTag:
		r.Seek(4, io.SeekCurrent)
		typeStr = "float32"

	case doubleTag:
		r.Seek(8, io.SeekCurrent)
		typeStr = "float64"

	case byteArrayTag:
		ln := int64(readInt32(r))
		r.Seek(ln, io.SeekCurrent)
		typeStr = "[]byte"

	case stringTag:
		ln := int64(readUInt16(r))
		r.Seek(ln, io.SeekCurrent)
		typeStr = "string"

	case listTag:
		listType := readByte(r)
		ln := int(readInt32(r))

		typeStr = getFieldType(r, listType)
		for i := 1; i < ln; i++ {
			typeStr = getFieldType(r, listType)
		}
		typeStr = fmt.Sprintf("[]%s", typeStr)

	case compoundTag:
		var sb strings.Builder
		sb.WriteString(" struct {\n")
		for {
			tagType := readByte(r)
			if tagType == endTag {
				break
			}
			name := toPascalCase(getKey(r))
			typeStr = getFieldType(r, tagType)

			sb.WriteString(fmt.Sprintf(
				"%s %s\n",
				name,
				typeStr,
			))
		}
		sb.WriteString("}")
		typeStr = sb.String()

	case intArrayTag:
		ln := int64(readInt32(r))
		r.Seek(ln*4, io.SeekCurrent)
		typeStr = "[]int32"

	case longArrayTag:
		ln := int64(readInt32(r))
		r.Seek(ln*8, io.SeekCurrent)
		typeStr = "[]int64"

	case endTag:
		typeStr = "interface{}"
	}

	return typeStr
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
	if strings.TrimSpace(out) == "" {
		return "Unnamed"
	}
	return out
}

func getKey(r io.Reader) string {
	keyLen := int(readInt16(r))
	s := readString(r, keyLen)
	return s
}

func readByte(r io.Reader) byte {
	b := make([]byte, 1)
	r.Read(b)
	return b[0]
}

func readString(r io.Reader, length int) string {
	b := make([]byte, length)
	r.Read(b)
	return string(b)
}

func readInt16(r io.Reader) int16 {
	var n int16
	binary.Read(r, binary.BigEndian, &n)
	return n
}

func readUInt16(r io.Reader) uint16 {
	var n uint16
	binary.Read(r, binary.BigEndian, &n)
	return n
}

func readInt32(r io.Reader) int32 {
	var n int32
	binary.Read(r, binary.BigEndian, &n)
	return n
}

func readInt64(r io.Reader) int64 {
	var s int64
	binary.Read(r, binary.BigEndian, &s)
	return s
}
