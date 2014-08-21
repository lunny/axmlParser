package axmlParser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/go-xweb/log"
)

const (
	TAG = "CXP"

	WORD_START_DOCUMENT = 0x00080003

	WORD_STRING_TABLE = 0x001C0001
	WORD_RES_TABLE    = 0x00080180

	WORD_START_NS  = 0x00100100
	WORD_END_NS    = 0x00100101
	WORD_START_TAG = 0x00100102
	WORD_END_TAG   = 0x00100103
	WORD_TEXT      = 0x00100104
	WORD_EOS       = 0xFFFFFFFF
	WORD_SIZE      = 4

	TYPE_ID_REF   = 0x01000008
	TYPE_ATTR_REF = 0x02000008
	TYPE_STRING   = 0x03000008
	TYPE_DIMEN    = 0x05000008
	TYPE_FRACTION = 0x06000008
	TYPE_INT      = 0x10000008
	TYPE_FLOAT    = 0x04000008

	TYPE_FLAGS  = 0x11000008
	TYPE_BOOL   = 0x12000008
	TYPE_COLOR  = 0x1C000008
	TYPE_COLOR2 = 0x1D000008
)

var (
	DIMEN = []string{"px", "dp", "sp",
		"pt", "in", "mm"}
)

type Parser struct {
	// Data
	listener Listener

	// Internal
	Namespaces map[string]string
	Data       []byte

	StringsTable                        []string
	ResourcesIds                        []int
	StringsCount, StylesCount, ResCount int
	ParserOffset                        int
}

func New(listener Listener) *Parser {
	return &Parser{
		listener:     listener,
		Namespaces:   make(map[string]string),
		Data:         make([]byte, 0),
		StringsTable: make([]string, 0),
		ResourcesIds: make([]int, 0),
	}
}

func (parser *Parser) IsValid(header []byte) bool {
	return (header[0] == 0x03) && (header[1] == 0x00) &&
		(header[2] == 0x08) && (header[3] == 0x00)
}

func (parser *Parser) Parse(data []byte) error {
	var word0 int64
	parser.Data = data

	for parser.ParserOffset < len(parser.Data) {
		word0 = int64(parser.getLEWord(parser.ParserOffset))
		switch word0 {
		case WORD_START_DOCUMENT:
			parser.parseStartDocument()
		case WORD_STRING_TABLE:
			parser.parseStringTable()
		case WORD_RES_TABLE:
			parser.parseResourceTable()
		case WORD_START_NS:
			parser.parseNamespace(true)
		case WORD_END_NS:
			parser.parseNamespace(false)
		case WORD_START_TAG:
			parser.parseStartTag()
		case WORD_END_TAG:
			parser.parseEndTag()
		case WORD_TEXT:
			parser.parseText()
		case WORD_EOS:
			parser.listener.EndDocument()
		default:
			parser.ParserOffset += WORD_SIZE
			log.Warnf(TAG+"Unknown word 0x%x @", word0, parser.ParserOffset)
		}
	}

	parser.listener.EndDocument()
	return nil
}

/**
 * A doc starts with the following 4bytes words :
 * <ul>
 * <li>0th word : 0x00080003</li>
 * <li>1st word : chunk size</li>
 * </ul>
 */
func (parser *Parser) parseStartDocument() {
	parser.listener.StartDocument()
	parser.ParserOffset += (2 * WORD_SIZE)
}

/**
 * the string table starts with the following 4bytes words :
 * <ul>
 * <li>0th word : 0x1c0001</li>
 * <li>1st word : chunk size</li>
 * <li>2nd word : number of string in the string table</li>
 * <li>3rd word : number of styles in the string table</li>
 * <li>4th word : ???? (0)</li>
 * <li>5th word : Offset to String data</li>
 * <li>6th word : Offset to style data</li>
 * </ul>
 */
func (parser *Parser) parseStringTable() {
	chunk := parser.getLEWord(parser.ParserOffset + (1 * WORD_SIZE))
	parser.StringsCount = parser.getLEWord(parser.ParserOffset + (2 * WORD_SIZE))
	parser.StylesCount = parser.getLEWord(parser.ParserOffset + (3 * WORD_SIZE))
	strOffset := parser.ParserOffset + parser.getLEWord(parser.ParserOffset+(5*WORD_SIZE))
	styleOffset := parser.getLEWord(parser.ParserOffset + (6 * WORD_SIZE))

	parser.StringsTable = make([]string, parser.StringsCount)
	var offset int
	for i := 0; i < int(parser.StringsCount); i++ {
		offset = strOffset + parser.getLEWord(parser.ParserOffset+((i+7)*WORD_SIZE))
		parser.StringsTable[i] = parser.getStringFromStringTable(offset)
	}

	if styleOffset > 0 {
		log.Warn(TAG, "Unread styles")
		for i := 0; i < parser.StylesCount; i++ {
			// TODO read the styles ???
		}
	}

	parser.ParserOffset += chunk
}

/**
 * the resource ids table starts with the following 4bytes words :
 * <ul>
 * <li>0th word : 0x00080180</li>
 * <li>1st word : chunk size</li>
 * </ul>
 */
func (parser *Parser) parseResourceTable() {
	chunk := parser.getLEWord(parser.ParserOffset + (1 * WORD_SIZE))
	parser.ResCount = (chunk / 4) - 2

	parser.ResourcesIds = make([]int, parser.ResCount)
	for i := 0; i < parser.ResCount; i++ {
		parser.ResourcesIds[i] = parser.getLEWord(parser.ParserOffset + ((i + 2) * WORD_SIZE))
	}

	parser.ParserOffset += chunk
}

/**
 * A namespace tag contains the following 4bytes words :
 * <ul>
 * <li>0th word : 0x00100100 = Start NS / 0x00100101 = end NS</li>
 * <li>1st word : chunk size</li>
 * <li>2nd word : line this tag appeared</li>
 * <li>3rd word : ??? (always 0xFFFFFF)</li>
 * <li>4th word : index of namespace prefix in StringIndexTable</li>
 * <li>5th word : index of namespace uri in StringIndexTable</li>
 * </ul>
 */
func (parser *Parser) parseNamespace(start bool) {
	prefixIdx := parser.getLEWord(parser.ParserOffset + (4 * WORD_SIZE))
	uriIdx := parser.getLEWord(parser.ParserOffset + (5 * WORD_SIZE))

	uri := parser.getString(uriIdx)
	prefix := parser.getString(prefixIdx)

	if start {
		parser.listener.StartPrefixMapping(prefix, uri)
		parser.Namespaces[uri] = prefix
	} else {
		parser.listener.EndPrefixMapping(prefix, uri)
		delete(parser.Namespaces, uri)
	}

	// Offset to first tag
	parser.ParserOffset += (6 * WORD_SIZE)
}

/**
 * A start tag will start with the following 4bytes words :
 * <ul>
 * <li>0th word : 0x00100102 = Start_Tag</li>
 * <li>1st word : chunk size</li>
 * <li>2nd word : line this tag appeared in the original file</li>
 * <li>3rd word : ??? (always 0xFFFFFF)</li>
 * <li>4th word : index of namespace uri in StringIndexTable, or 0xFFFFFFFF
 * for default NS</li>
 * <li>5th word : index of element name in StringIndexTable</li>
 * <li>6th word : ???</li>
 * <li>7th word : number of attributes following the start tag</li>
 * <li>8th word : ??? (0)</li>
 * </ul>
 *
 */
func (parser *Parser) parseStartTag() {
	// get tag info
	uriIdx := parser.getLEWord(parser.ParserOffset + (4 * WORD_SIZE))
	nameIdx := parser.getLEWord(parser.ParserOffset + (5 * WORD_SIZE))
	attrCount := parser.getLEWord(parser.ParserOffset + (7 * WORD_SIZE))

	name := parser.getString(nameIdx)
	var uri, qname string
	if int64(uriIdx) == 0xFFFFFFFF {
		uri = ""
		qname = name
	} else {
		uri = parser.getString(uriIdx)
		if v, ok := parser.Namespaces[uri]; ok {
			qname = v + ":" + name
		} else {
			qname = name
		}
	}

	// offset to start of attributes
	parser.ParserOffset += (9 * WORD_SIZE)

	attrs := make([]*Attribute, attrCount) // NOPMD
	for a := 0; a < attrCount; a++ {
		attrs[a] = parser.parseAttribute() // NOPMD

		// offset to next attribute or tag
		parser.ParserOffset += (5 * 4)
	}

	parser.listener.StartElement(uri, name, qname, attrs)
}

/**
 * An attribute will have the following 4bytes words :
 * <ul>
 * <li>0th word : index of namespace uri in StringIndexTable, or 0xFFFFFFFF
 * for default NS</li>
 * <li>1st word : index of attribute name in StringIndexTable</li>
 * <li>2nd word : index of attribute value, or 0xFFFFFFFF if value is a
 * typed value</li>
 * <li>3rd word : value type</li>
 * <li>4th word : resource id value</li>
 * </ul>
 */
func (parser *Parser) parseAttribute() *Attribute {
	attrNSIdx := parser.getLEWord(parser.ParserOffset)
	attrNameIdx := parser.getLEWord(parser.ParserOffset + (1 * WORD_SIZE))
	attrValueIdx := parser.getLEWord(parser.ParserOffset + (2 * WORD_SIZE))
	attrType := parser.getLEWord(parser.ParserOffset + (3 * WORD_SIZE))
	attrData := parser.getLEWord(parser.ParserOffset + (4 * WORD_SIZE))

	attr := new(Attribute)
	attr.Name = parser.getString(attrNameIdx)

	if int64(attrNSIdx) == 0xFFFFFFFF {
		attr.Namespace = ""
		attr.Prefix = ""
	} else {
		uri := parser.getString(attrNSIdx)
		if v, ok := parser.Namespaces[uri]; ok {
			attr.Namespace = uri
			attr.Prefix = v
		}
	}

	if int64(attrValueIdx) == 0xFFFFFFFF {
		attr.Value = parser.getAttributeValue(attrType, attrData)
	} else {
		attr.Value = parser.getString(attrValueIdx)
	}

	return attr
}

/**
 * A text will start with the following 4bytes word :
 * <ul>
 * <li>0th word : 0x00100104 = Text</li>
 * <li>1st word : chunk size</li>
 * <li>2nd word : line this element appeared in the original document</li>
 * <li>3rd word : ??? (always 0xFFFFFFFF)</li>
 * <li>4rd word : string index in string table</li>
 * <li>5rd word : ??? (always 8)</li>
 * <li>6rd word : ??? (always 0)</li>
 * </ul>
 *
 */
func (parser *Parser) parseText() {
	// get tag infos
	strIndex := parser.getLEWord(parser.ParserOffset + (4 * WORD_SIZE))

	data := parser.getString(strIndex)
	parser.listener.CharacterData(data)

	// offset to next node
	parser.ParserOffset += (7 * WORD_SIZE)
}

/**
 * EndTag contains the following 4bytes words :
 * <ul>
 * <li>0th word : 0x00100103 = End_Tag</li>
 * <li>1st word : chunk size</li>
 * <li>2nd word : line this tag appeared in the original file</li>
 * <li>3rd word : ??? (always 0xFFFFFFFF)</li>
 * <li>4th word : index of namespace name in StringIndexTable, or 0xFFFFFFFF
 * for default NS</li>
 * <li>5th word : index of element name in StringIndexTable</li>
 * </ul>
 */
func (parser *Parser) parseEndTag() {
	// get tag info
	uriIdx := parser.getLEWord(parser.ParserOffset + (4 * WORD_SIZE))
	nameIdx := parser.getLEWord(parser.ParserOffset + (5 * WORD_SIZE))

	name := parser.getString(nameIdx)
	var uri string
	if int64(uriIdx) == 0xFFFFFFFF {
		uri = ""
	} else {
		uri = parser.getString(uriIdx)
	}

	parser.listener.EndElement(uri, name, "")

	// offset to start of next tag
	parser.ParserOffset += (6 * WORD_SIZE)
}

/**
 * @param index
 *            the index of the string in the StringIndexTable
 * @return the string
 */
func (parser *Parser) getString(index int) string {
	var res string
	if (index >= 0) && (index < parser.StringsCount) {
		res = parser.StringsTable[index]
	} else {
		res = "" // NOPMD
	}

	return res
}

/**
 * @param offset
 *            offset of the beginning of the string inside the StringTable
 *            (and not the whole data array)
 * @return the String
 */
func (parser *Parser) getStringFromStringTable(offset int) string {
	var strLength int
	var chars []byte
	if parser.Data[offset+1] == parser.Data[offset] {
		strLength = int(parser.Data[offset])
		chars = make([]byte, strLength) // NOPMD
		for i := 0; i < strLength; i++ {
			chars[i] = parser.Data[offset+2+i] // NOPMD
		}
	} else {
		strLength = ((int(parser.Data[offset+1] << 8)) & 0xFF00) |
			(int(parser.Data[offset]) & 0xFF)
		chars = make([]byte, strLength) // NOPMD
		for i := 0; i < strLength; i++ {
			chars[i] = parser.Data[offset+2+(i*2)] // NOPMD
		}
	}
	return string(chars)
}

/**
 * @param arr
 *            the byte array to read
 * @param off
 *            the offset of the word to read
 * @return value of a Little Endian 32 bit word from the byte arrayat offset
 *         off.
 */
func (parser *Parser) getLEWord(off int) int {
	return int(int((int64(parser.Data[off+3])<<24)&0xff000000) |
		((int(parser.Data[off+2]) << 16) & 0x00ff0000) |
		((int(parser.Data[off+1]) << 8) & 0x0000ff00) |
		((int(parser.Data[off+0]) << 0) & 0x000000ff))
}

/**
 * @param word
 *            a word read in an attribute data
 * @return the typed value
 */
func (parser *Parser) getAttributeValue(tpe int, data int) string {
	var res string

	switch tpe {
	case TYPE_STRING:
		res = parser.getString(data)
	case TYPE_DIMEN:
		res = fmt.Sprintf("%v", data>>8) + DIMEN[data&0xFF]
	case TYPE_FRACTION:
		fracValue := (float64(data) / (float64(0x7FFFFFFF)))
		// res = String.format("%.2f%%", fracValue);
		//res = new DecimalFormat("#.##%").format(fracValue)
		res = fmt.Sprintf("%.2f%%", fracValue)
	case TYPE_FLOAT:
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, int32(data))
		if err != nil {
			panic(err)
		}
		var fl float32
		err = binary.Read(buf, binary.LittleEndian, &fl)
		if err != nil {
			panic(err)
		}
		//res = Float.toString(Float.intBitsToFloat(data))
		res = fmt.Sprintf("%f", fl)
	case TYPE_INT:
		fallthrough
	case TYPE_FLAGS:
		res = strconv.Itoa(data)
	case TYPE_BOOL:
		//res = Boolean.toString(data != 0)
		if data != 0 {
			res = "true"
		} else {
			res = "false"
		}
	case TYPE_COLOR:
		fallthrough
	case TYPE_COLOR2:
		res = fmt.Sprintf("%#08X", data)
	case TYPE_ID_REF:
		res = fmt.Sprintf("@id/0x%08X", data)
	case TYPE_ATTR_REF:
		res = fmt.Sprintf("?id/0x%08X", data)
	default:
		log.Warnf(TAG+"(type=%d) : %v (0x%08X) @%d", tpe, data,
			data, parser.ParserOffset)
		res = fmt.Sprintf("%08X/0x%08X", tpe, data)
	}

	return res
}
