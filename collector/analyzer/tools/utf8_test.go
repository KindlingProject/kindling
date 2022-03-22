package tools

import (
	"testing"
	"unicode/utf8"
)

func TestFomratByteArrayToUtf8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  string
		valid bool
	}{
		{name: "normal", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}, want: "Hello, 世界", valid: true},
		{name: "substring1", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95}, want: "Hello, 世"},
		{name: "substring2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7}, want: "Hello, 世"},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96}, want: "Hello, 世", valid: true},
		{name: "substring4", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8}, want: "Hello, "},
		{name: "substring5", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4}, want: "Hello, "},
		{name: "substring6", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' '}, want: "Hello, ", valid: true},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ','}, want: "Hello,", valid: true},
		{name: "invalid", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 'e', 0xe7, 0xe4, 0x95}, want: "Hello, 世e"},
		{name: "invalid2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 'e', 0xe4, 0x95}, want: "Hello, 世"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FomratByteArrayToUtf8(tt.data)
			if !utf8.ValidString(got) {
				t.Errorf("Fail to format byteArray to utf8")
			}

			if utf8.Valid(tt.data) != tt.valid {
				t.Errorf("Fail to check validte, got = %v, want %v", !tt.valid, tt.valid)
			}

			if got != tt.want {
				t.Errorf("Fail to check value, got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFomratStringToUtf8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  string
		valid bool
	}{
		{name: "normal", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}, want: "Hello, 世界", valid: true},
		{name: "substring1", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95}, want: "Hello, 世"},
		{name: "substring2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7}, want: "Hello, 世"},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96}, want: "Hello, 世", valid: true},
		{name: "substring4", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8}, want: "Hello, "},
		{name: "substring5", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4}, want: "Hello, "},
		{name: "substring6", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' '}, want: "Hello, ", valid: true},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ','}, want: "Hello,", valid: true},
		{name: "invalid", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 'e', 0xe7, 0xe4, 0x95}, want: "Hello, 世e"},
		{name: "invalid2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 'e', 0xe4, 0x95}, want: "Hello, 世"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := string(tt.data)
			got := FomratStringToUtf8(data)
			if !utf8.ValidString(got) {
				t.Errorf("Fail to format byteArray to utf8")
			}

			if utf8.ValidString(data) != tt.valid {
				t.Errorf("Fail to check validte, got = %v, want %v", !tt.valid, tt.valid)
			}

			if got != tt.want {
				t.Errorf("Fail to check value, got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	datas := []string{
		"",
		"a",
		"abc",
		"Ж",
		"ЖЖ",
		"брэд-ЛГТМ",
		"☺☻☹",
		"aa\xe2",
		string([]byte{66, 250}),
		string([]byte{66, 250, 67}),
		"a\uFFFDb",
		string("\xF4\x8F\xBF\xBF"),     // U+10FFFF
		string("\xF4\x90\x80\x80"),     // U+10FFFF+1; out of range
		string("\xF7\xBF\xBF\xBF"),     // 0x1FFFFF; out of range
		string("\xFB\xBF\xBF\xBF\xBF"), // 0x3FFFFFF; out of range
		string("\xc0\x80"),             // U+0000 encoded in two bytes: incorrect
		string("\xed\xa0\x80"),         // U+D800 high surrogate (sic)
		string("\xed\xbf\xbf"),         // U+DFFF low surrogate (sic)
	}

	for _, data := range datas {
		got := FomratStringToUtf8(data)
		if !utf8.ValidString(got) {
			t.Errorf("Fail to format to utf8: %s", data)
		}
	}
}
