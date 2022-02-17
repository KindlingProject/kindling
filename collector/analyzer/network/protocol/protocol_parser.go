package protocol

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/model"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	PARSE_FAIL     = 0
	PARSE_OK       = 1
	PARSE_COMPLETE = 2

	EOF = -1
)

var (
	ErrMessageComplete = errors.New("message completed")
	ErrMessageShort    = errors.New("message is too short")
	ErrMessageInvalid  = errors.New("message is invalid")
	ErrEof             = errors.New("EOF")
	ErrUnexpectedEOF   = errors.New("unexpected EOF")
)

type PayloadMessage struct {
	Data         []byte
	Offset       int
	attributeMap *model.AttributeMap
}

func NewRequestMessage(data []byte) *PayloadMessage {
	return &PayloadMessage{
		Data:         data,
		Offset:       0,
		attributeMap: model.NewAttributeMap(),
	}
}

func NewResponseMessage(data []byte, attributeMap *model.AttributeMap) *PayloadMessage {
	return &PayloadMessage{
		Data:         data,
		Offset:       0,
		attributeMap: attributeMap,
	}
}

func (message *PayloadMessage) IsComplete() bool {
	return len(message.Data) <= message.Offset
}

func (message *PayloadMessage) HasMoreLength(length int) bool {
	return message.Offset+length <= len(message.Data)
}

func (message *PayloadMessage) GetData(offset int, length int) []byte {
	if offset+length > len(message.Data) {
		return message.Data[offset:]
	}
	return message.Data[offset : offset+length]
}

// =============== Attributes ===============
func (message PayloadMessage) GetAttributes() *model.AttributeMap {
	return message.attributeMap
}

func (message PayloadMessage) AddIntAttribute(key string, value int64) {
	message.attributeMap.AddIntValue(key, value)
}

func (message PayloadMessage) AddStringAttribute(key string, value string) {
	message.attributeMap.AddStringValue(key, value)
}

func (message PayloadMessage) AddBoolAttribute(key string, value bool) {
	message.attributeMap.AddBoolValue(key, value)
}

func (message PayloadMessage) GetIntAttribute(key string) int64 {
	return message.attributeMap.GetIntValue(key)
}

func (message PayloadMessage) GetStringAttribute(key string) string {
	return message.attributeMap.GetStringValue(key)
}

func (message PayloadMessage) GetBoolAttribute(key string) bool {
	return message.attributeMap.GetBoolValue(key)
}

func (message PayloadMessage) HasAttribute(key string) bool {
	return message.attributeMap.HasAttribute(key)
}

// =============== PayLoad ===============
func (message *PayloadMessage) ReadLittleEndianInt(offset int, size int) int32 {
	if size == 1 {
		return int32(message.Data[offset])
	}
	if size == 2 {
		return int32(message.Data[offset]) | int32(message.Data[offset+1])<<8
	}
	if size == 3 {
		return int32(message.Data[offset]) | int32(message.Data[offset+1])<<8 | int32(message.Data[offset+2])<<16
	}
	return int32(message.Data[offset]) | int32(message.Data[offset+1])<<8 | int32(message.Data[offset+2])<<16 | int32(message.Data[offset+2])<<24
}

func (message *PayloadMessage) ReadUnsignedIntByThreeBytes(offset int) uint32 {
	return (uint32(message.Data[offset])<<16 | uint32(message.Data[offset+1])<<8 | uint32(message.Data[offset+2]))
}

func (message *PayloadMessage) ReadUInt16(offset int) (complete bool, value uint16) {
	if offset+2 >= len(message.Data) {
		return true, 0
	}
	return false, uint16(message.Data[offset])<<8 | uint16(message.Data[offset+1])
}

func (message *PayloadMessage) ReadInt16(offset int, v *int16) (toOffset int, err error) {
	if offset < -1 {
		return -1, ErrMessageInvalid
	}
	if offset+2 >= len(message.Data) {
		return -1, ErrMessageShort
	}
	*v = int16(message.Data[offset])<<8 | int16(message.Data[offset+1])
	return offset + 2, nil
}

func (message *PayloadMessage) ReadInt32(offset int, v *int32) (toOffset int, err error) {
	if offset < -1 {
		return -1, ErrMessageInvalid
	}
	if offset+4 >= len(message.Data) {
		return -1, ErrMessageShort
	}
	*v = int32(message.Data[offset])<<24 | int32(message.Data[offset+1])<<16 | int32(message.Data[offset+2])<<8 | int32(message.Data[offset+3])
	return offset + 4, nil
}

func (message *PayloadMessage) ReadThreeBytes(offset int) int32 {
	return int32(message.Data[offset+2])<<16 |
		int32(message.Data[offset+1])<<8 | int32(message.Data[offset])
}

func (message *PayloadMessage) ReadInt64(offset int) int64 {
	return int64(message.Data[offset])<<56 |
		int64(message.Data[offset+1])<<48 |
		int64(message.Data[offset+2])<<40 |
		int64(message.Data[offset+3])<<32 |
		int64(message.Data[offset+4])<<24 |
		int64(message.Data[offset+5])<<16 |
		int64(message.Data[offset+6])<<8 |
		int64(message.Data[offset+7])
}

func (message *PayloadMessage) ReadByte() (byte, error) {
	if message.Offset >= len(message.Data) {
		return 0, ErrMessageComplete
	}
	message.Offset += 1
	return message.Data[message.Offset-1], nil
}

func (message *PayloadMessage) ReadFull(buf []byte) (n int, err error) {
	var min = len(buf)
	if len(message.Data)-message.Offset < min {
		return 0, ErrMessageShort
	}
	for n < min && err == nil {
		var nn int
		nn, err = message.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == ErrEof {
		err = ErrUnexpectedEOF
	}
	return
}

func (message *PayloadMessage) Read(b []byte) (n int, err error) {
	if message.Offset >= len(message.Data) {
		return 0, ErrEof
	}
	n = copy(b, message.Data[message.Offset:])
	message.Offset = message.Offset + n
	return
}

func (message *PayloadMessage) ReadBytes(offset int, length int) (toOffset int, value []byte) {
	maxLength := offset + length
	if maxLength >= len(message.Data) {
		return EOF, nil
	}
	return maxLength, message.Data[offset:maxLength]
}

func (message *PayloadMessage) readUnsignedVarIntCore(offset int, times int, f func(uint64)) (toOffset int, err error) {
	var b byte
	x := uint64(0)
	s := uint(0)
	for i := 0; i < times; i++ {
		if offset+i >= len(message.Data) {
			return -1, ErrMessageShort
		}
		b = message.Data[offset+i]
		if b < 0x80 {
			x |= uint64(b) << s
			f(x)
			return offset + i + 1, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return -1, ErrMessageInvalid
}

func (message *PayloadMessage) ReadUnsignedVarInt(offset int, v *uint64) (toOffset int, err error) {
	return message.readUnsignedVarIntCore(offset, 5, func(value uint64) { *v = value })
}

func (message *PayloadMessage) ReadVarInt(offset int, v *int64) (toOffset int, err error) {
	return message.readUnsignedVarIntCore(offset, 5, func(value uint64) { *v = int64(value>>1) ^ -(int64(value) & 1) })
}

func (message *PayloadMessage) ReadNullableString(offset int, compact bool, v *string) (toOffset int, err error) {
	if compact {
		return message.readCompactNullableString(offset, v)
	}
	return message.readNullableString(offset, v)
}

func (message *PayloadMessage) readNullableString(offset int, v *string) (toOffset int, err error) {
	var length int16
	if toOffset, err = message.ReadInt16(offset, &length); err != nil {
		return toOffset, err
	}
	if length < -1 {
		return -1, ErrMessageInvalid
	}
	if length == -1 {
		return toOffset, nil
	}
	*v = string(message.Data[toOffset : toOffset+int(length)])
	return toOffset + int(length), nil
}

func (message *PayloadMessage) readCompactNullableString(offset int, v *string) (toOffset int, err error) {
	var length uint64
	if toOffset, err = message.ReadUnsignedVarInt(offset, &length); err != nil {
		return toOffset, err
	}
	intLength := int(length)
	intLength -= 1
	if intLength < -1 {
		return -1, ErrMessageInvalid
	}
	if intLength == -1 {
		return toOffset, nil
	}
	*v = string(message.Data[toOffset : toOffset+intLength])
	return toOffset + intLength, nil
}

func (message *PayloadMessage) ReadArraySize(offset int, compact bool, size *int32) (toOffset int, err error) {
	if compact {
		return message.readCompactArraySize(offset, size)
	}
	return message.readArraySize(offset, size)
}

func (message *PayloadMessage) readCompactArraySize(offset int, size *int32) (toOffset int, err error) {
	var length uint64
	if toOffset, err = message.ReadUnsignedVarInt(offset, &length); err != nil {
		return toOffset, err
	}
	len := int32(length)
	if len < 0 {
		return -1, ErrMessageInvalid
	}
	if len == 0 {
		*size = 0
		return toOffset, nil
	}
	len -= 1
	*size = len
	return toOffset, nil
}

func (message *PayloadMessage) readArraySize(offset int, size *int32) (toOffset int, err error) {
	var length int32
	if toOffset, err = message.ReadInt32(offset, &length); err != nil {
		return toOffset, err
	}
	if length < -1 {
		return -1, ErrMessageInvalid
	}
	if length == -1 {
		*size = 0
		return toOffset, nil
	}
	*size = length
	return toOffset, nil
}

func (message *PayloadMessage) ReadString(offset int, compact bool, v *string) (toOffset int, err error) {
	if compact {
		return message.readCompactString(offset, v)
	}
	return message.readString(offset, v)
}

func (message *PayloadMessage) readCompactString(offset int, v *string) (toOffset int, err error) {
	var length uint64
	if toOffset, err = message.ReadUnsignedVarInt(offset, &length); err != nil {
		return toOffset, err
	}

	intLen := int(length)
	intLen -= 1
	if intLen < 0 {
		return -1, ErrMessageInvalid
	}
	if toOffset+int(length) >= len(message.Data) {
		*v = string(message.Data[toOffset:])
		return len(message.Data), nil
	}
	*v = string(message.Data[toOffset : toOffset+intLen])
	return toOffset + intLen, nil
}

func (message *PayloadMessage) readString(offset int, v *string) (toOffset int, err error) {
	var length int16
	if toOffset, err = message.ReadInt16(offset, &length); err != nil {
		return toOffset, err
	}
	if length < 0 {
		return -1, ErrMessageInvalid
	}
	if toOffset+int(length) >= len(message.Data) {
		*v = string(message.Data[toOffset:])
		return len(message.Data), nil
	}

	*v = string(message.Data[toOffset : toOffset+int(length)])
	return toOffset + int(length), nil
}

func (message *PayloadMessage) ReadUntilBlank(from int) (int, []byte) {
	var length = len(message.Data)

	for i := from; i < length; i++ {
		if message.Data[i] == ' ' {
			return i + 1, message.Data[from:i]
		}
	}
	return length, message.Data[from:length]
}

func (message *PayloadMessage) ReadUntilBlankWithLength(from int, fixedLength int) (int, []byte) {
	var length = len(message.Data)
	if fixedLength+from < length {
		length = from + fixedLength
	}

	for i := from; i < length; i++ {
		if message.Data[i] == ' ' {
			return i + 1, message.Data[from:i]
		}
	}
	return length, message.Data[from:length]
}

// Read Util \r\n
func (message *PayloadMessage) ReadUntilCRLF(from int) (offset int, data []byte) {
	var length = len(message.Data)
	if from >= length {
		return EOF, nil
	}

	for i := from; i < length; i++ {
		if message.Data[i] != '\r' {
			continue
		}

		if i == length-1 {
			// End with \r
			offset = length
			data = message.Data[from : length-1]
			return
		} else if message.Data[i+1] == '\n' {
			// \r\n
			offset = i + 2
			data = message.Data[from:i]
			return
		} else {
			return EOF, nil
		}
	}

	offset = length
	data = message.Data[from:]
	return
}

type FastFailFn func(message *PayloadMessage) bool
type ParsePkgFn func(message *PayloadMessage) (bool, bool)
type PairMatch func(requests []*PayloadMessage, response *PayloadMessage) int

type ProtocolParser struct {
	protocol       string
	multiFrames    bool
	requestParser  PkgParser
	responseParser PkgParser
	pairMatch      PairMatch
	portCounter    cmap.ConcurrentMap
}

func NewProtocolParser(protocol string, requestParser PkgParser, responseParser PkgParser, pairMatch PairMatch) *ProtocolParser {
	return &ProtocolParser{
		protocol:       protocol,
		requestParser:  requestParser,
		responseParser: responseParser,
		pairMatch:      pairMatch,
		portCounter:    cmap.New(),
	}
}

func (parser *ProtocolParser) EnableMultiFrame() {
	parser.multiFrames = true
}

func (parser *ProtocolParser) GetProtocol() string {
	return parser.protocol
}

func (parser *ProtocolParser) MultiRequests() bool {
	return parser.pairMatch != nil
}

func (parser *ProtocolParser) PairMatch(requests []*PayloadMessage, response *PayloadMessage) int {
	if parser.pairMatch == nil {
		return -1
	}
	return parser.pairMatch(requests, response)
}

func (parser *ProtocolParser) ParseRequest(message *PayloadMessage) bool {
	return parser.requestParser.parsePayload(parser.multiFrames, message)
}

func (parser *ProtocolParser) ParseResponse(message *PayloadMessage) bool {
	return parser.responseParser.parsePayload(parser.multiFrames, message)
}

type PkgParser struct {
	fastFail FastFailFn
	parser   ParsePkgFn
	children []*PkgParser
}

func CreatePkgParser(fastFail FastFailFn, parser ParsePkgFn) PkgParser {
	return PkgParser{
		fastFail: fastFail,
		parser:   parser,
		children: nil,
	}
}

func (parent *PkgParser) Add(fastFail FastFailFn, parser ParsePkgFn) *PkgParser {
	child := &PkgParser{
		fastFail: fastFail,
		parser:   parser,
	}
	parent.children = append(parent.children, child)
	return child
}

func (current PkgParser) parsePayload(multiFrames bool, message *PayloadMessage) bool {
	if multiFrames {
		for {
			status := current.parseOneFrame(message)
			if status != PARSE_OK {
				return status == PARSE_COMPLETE
			}
		}
	}
	return current.parseOneFrame(message) != PARSE_FAIL
}

func (current PkgParser) parseOneFrame(message *PayloadMessage) int {
	if current.fastFail(message) {
		return PARSE_FAIL
	}
	ok, complete := current.parser(message)
	if !ok {
		return PARSE_FAIL
	}
	if complete {
		return PARSE_COMPLETE
	}

	// Continue when true, false
	if current.children != nil {
		for _, child := range current.children {
			status := child.parseOneFrame(message)
			if status != PARSE_FAIL {
				// Return when subProtocol parser finished or success
				return status
			}
		}
		return PARSE_FAIL
	}
	return PARSE_OK
}

func (parser *ProtocolParser) AddPortCount(port uint32) uint32 {
	key := strconv.Itoa(int(port))
	if val, ok := parser.portCounter.Get(key); ok {
		return atomic.AddUint32(val.(*uint32), 1)
	} else {
		count := uint32(1)
		parser.portCounter.Set(key, &count)
		return count
	}
}

func (parser *ProtocolParser) ResetPort(port uint32) {
	key := strconv.Itoa(int(port))
	parser.portCounter.Remove(key)
}
