package main

import (
    "bufio"
    "compress/gzip"
    "fmt"
    "os"
    "io"
)

const (
    tagStructEnd = 0 // No name. Single zero byte.
    tagInt8 = 1 // A single signed byte (8 bits)
    tagInt16 = 2 // A signed short (16 bits, big endian)
    tagInt32 = 3 // A signed int (32 bits, big endian)
    tagInt64 = 4 // A signed long (64 bits, big endian)
    tagFloat32 = 5 // A floating point value (32 bits, big endian, IEEE 754-2008, binary32)
    tagFloat64 = 6 // A floating point value (64 bits, big endian, IEEE 754-2008, binary64)
    tagByteArray = 7 // { TAG_Int length; An array of bytes of unspecified format. The length of this array is <length> bytes }
    tagString = 8 // { TAG_Short length; An array of bytes defining a string in UTF-8 format. The length of this array is <length> bytes }
    tagList = 9 // { TAG_Byte tagId; TAG_Int length; A sequential list of Tags (not Named Tags), of type <typeId>. The length of this array is <length> Tags. } Notes: All tags share the same type.
    tagStruct = 10 // { A sequential list of Named Tags. This array keeps going until a TAG_End is found.; TAG_End end } Notes: If there's a nested TAG_Compound within this tag, that one will also have a TAG_End, so simply reading until the next TAG_End will not work. The names of the named tags have to be unique within each TAG_Compound The order of the tags is not guaranteed.
)

func main() {
    var file, fileErr = os.Open("c.4.4.dat", os.O_RDONLY, 0666) // x: 76, y: 69
    if fileErr != nil {
        fmt.Println(fileErr)
        return
    }

    var r, rErr = gzip.NewReader(file)
    if rErr != nil {
        fmt.Println(rErr)
        return
    }

    var br = bufio.NewReader(r)

    for {
        var typeId, name, err = readTag(br)
        if err != nil {
            break
        }

        switch typeId {
        case tagStruct:
            fmt.Printf("%s struct start\n", name)
        case tagStructEnd:
            fmt.Printf("struct end\n")
        case tagByteArray:
            var bytes, err2 = readBytes(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            fmt.Printf("%s bytes(%d) %v\n", name, len(bytes), bytes)
            if name == "Blocks" {
                for i := 0; i < len(bytes); i += 128 {
                    var column = bytes[i:i+128]
                    for _, blockId := range column {
                        fmt.Printf("%2x", blockId)
                    }
                    fmt.Println()
                }
            }
        case tagInt8:
            var number, err2 = readInt8(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            fmt.Printf("%s int8 %v\n", name, number)
        case tagInt16:
            var number, err2 = readInt16(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            fmt.Printf("%s int16 %v\n", name, number)
        case tagInt32:
            var number, err2 = readInt32(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            fmt.Printf("%s int32 %v\n", name, number)
        case tagInt64:
            var number, err2 = readInt64(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            fmt.Printf("%s int64 %v\n", name, number)
        case tagList:
            var itemTypeId, length, err2 = readListHeader(br)
            if err2 != nil {
                fmt.Println(err2)
            }
            switch itemTypeId {
                case tagInt8:
                    fmt.Printf("%s list int8 (%d)\n", name, length)
                    for i := 0; i < length; i++ {
                        var v, err3 = readInt8(br)
                        if err3 != nil {
                            fmt.Println(err3)
                        }
                        fmt.Println("  ", v)
                    }
                default:
                    fmt.Printf("list todo(%n) %n\n", itemTypeId, length)
            }
        default:
            fmt.Printf("%s todo(%d)\n", name, typeId)
        }
    }
    fmt.Println()
}

func readTag(r *bufio.Reader) (byte, string, os.Error) {
    var typeId, err = r.ReadByte()
    if err != nil || typeId == 0 {
        return typeId, "", err
    }

    var name, err2 = readString(r)
    if err2 != nil {
        return typeId, name, err2
    }

    return typeId, name, nil
}

func readListHeader(r *bufio.Reader) (itemTypeId byte, length int, err os.Error) {
    length = 0

    itemTypeId, err = r.ReadByte()
    if err == nil {
        length, err = readInt32(r)
    }

    return
}

func readString(r *bufio.Reader) (string, os.Error) {
    var length, err1 = readInt16(r)
    if err1 != nil {
        return "", err1
    }

    var bytes = make([]byte, length)
    var _, err2 = io.ReadFull(r, bytes)
    return string(bytes), err2
}

func readBytes(r *bufio.Reader) ([]byte, os.Error) {
    var length, err1 = readInt32(r)
    if err1 != nil {
        return nil, err1
    }

    var bytes = make([]byte, length)
    var _, err2 = io.ReadFull(r, bytes)
    return bytes, err2
}

func readInt8(r *bufio.Reader) (int, os.Error) {
    return readIntN(r, 1)
}

func readInt16(r *bufio.Reader) (int, os.Error) {
    return readIntN(r, 2)
}

func readInt32(r *bufio.Reader) (int, os.Error) {
    return readIntN(r, 4)
}

func readInt64(r *bufio.Reader) (int, os.Error) {
    return readIntN(r, 8)
}

func readIntN(r *bufio.Reader, n int) (int, os.Error) {
    var a int = 0

    for i := 0; i < n; i++ {
        var b, err = r.ReadByte()
        if err != nil {
            return a, err
        }
        a = a << 8 + int(b)
    }

    return a, nil
}