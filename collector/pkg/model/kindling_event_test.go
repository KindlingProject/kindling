package model

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalEvent(t *testing.T) {
	v, _ := base64.StdEncoding.DecodeString("sd4btBwgHxc=")
	att1 := &KeyValue{
		Key:       "start_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666085694803271345), att1.GetUintValue())
	v, _ = base64.StdEncoding.DecodeString("19uyLHcgHxc=")
	att2 := &KeyValue{
		Key:       "end_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666086083373489111), att2.GetUintValue())
}
