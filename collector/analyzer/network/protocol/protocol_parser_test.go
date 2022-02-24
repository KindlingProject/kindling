package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadInt16(t *testing.T) {
	// ff 0 4 t e s t
	data := []byte{0xff, 0x00, 0x04, 0x74, 0x65, 0x73, 0x74}
	message := NewRequestMessage(data)

	tests := []struct {
		name   string
		offset int
		expect int16
		err    error
	}{
		{"Invalid Index", -1, -1, ErrMessageInvalid},
		{"Negative Number", 0, -256, nil},
		{"Positive Integer", 1, 4, nil},
		{"Large Integer", 2, 1140, nil},
		{"Overflow Index", 10, -1, ErrMessageShort},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var realValue int16
			if _, err := message.ReadInt16(test.offset, &realValue); err != nil {
				assert.Equal(t, test.err, err)
				return
			}
			assert.Equal(t, test.expect, realValue)
		})
	}
}

func TestReadInt32(t *testing.T) {
	// ff 0 0 0 4 t e s t
	data := []byte{0xff, 0x00, 0x00, 0x00, 0x04, 0x74, 0x65, 0x73, 0x74}
	message := NewRequestMessage(data)

	tests := []struct {
		name   string
		offset int
		expect int32
		err    error
	}{
		{"Invalid Index", -1, -1, ErrMessageInvalid},
		{"Negative Number", 0, -16777216, nil},
		{"Positive Integer", 1, 4, nil},
		{"Large Integer", 2, 1140, nil},
		{"Overflow Index", 10, -1, ErrMessageShort},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var realValue int32
			if _, err := message.ReadInt32(test.offset, &realValue); err != nil {
				assert.Equal(t, test.err, err)
				return
			}
			assert.Equal(t, test.expect, realValue)
		})
	}
}

func TestReadNullableString(t *testing.T) {
	// ff 0 4 t e s t
	data := []byte{0xff, 0x00, 0x04, 0x74, 0x65, 0x73, 0x74}
	message := NewRequestMessage(data)

	tests := []struct {
		name   string
		offset int
		expect string
		err    error
	}{
		{"Invalid Index", -1, "", ErrMessageInvalid},
		{"Invalid Length", 0, "", ErrMessageInvalid},
		{"Normal String", 1, "test", nil},
		{"Trim String", 2, "est", nil},
		{"Overflow Index", 10, "", ErrMessageShort},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var realValue string
			if _, err := message.ReadNullableString(test.offset, false, &realValue); err != nil {
				assert.Equal(t, test.err, err)
				return
			}
			assert.Equal(t, test.expect, realValue)
		})
	}
}
