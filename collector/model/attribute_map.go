package model

import (
	"encoding/json"
	"strconv"
)

type AttributeMap struct {
	values map[string]AttributeValue
}

func NewAttributeMap() *AttributeMap {
	values := make(map[string]AttributeValue)
	return &AttributeMap{values}
}

func NewAttributeMapWithValues(values map[string]AttributeValue) *AttributeMap {
	return &AttributeMap{values: values}
}

func NewStringValue(value string) AttributeValue {
	return &stringValue{value: value}
}

func NewIntValue(value int64) AttributeValue {
	return &intValue{value: value}
}

func NewBoolValue(value bool) AttributeValue {
	return &boolValue{value: value}
}

func (attributes *AttributeMap) Merge(other *AttributeMap) {
	if other == nil {
		return
	}
	for k, v := range other.values {
		attributes.values[k] = v
	}
}

func (attributes *AttributeMap) Size() int {
	return len(attributes.values)
}

func (attributes *AttributeMap) IsEmpty() bool {
	return len(attributes.values) == 0
}

func (attributes *AttributeMap) HasAttribute(key string) bool {
	_, existing := attributes.values[key]
	return existing
}

func (attributes *AttributeMap) GetStringValue(key string) string {
	value := attributes.values[key]
	if x, ok := value.(*stringValue); ok {
		return x.value
	}
	return ""
}

func (attributes *AttributeMap) AddStringValue(key string, value string) {
	attributes.values[key] = &stringValue{
		value: value,
	}
}

func (attributes *AttributeMap) UpdateAddStringValue(key string, value string) {
	if v, ok := attributes.values[key]; ok {
		v.(*stringValue).value = value
	} else {
		attributes.AddStringValue(key, value)
	}
}

func (attributes *AttributeMap) GetIntValue(key string) int64 {
	value := attributes.values[key]
	if x, ok := value.(*intValue); ok {
		return x.value
	}
	return 0
}

func (attributes *AttributeMap) AddIntValue(key string, value int64) {
	attributes.values[key] = &intValue{
		value: value,
	}
}

func (attributes *AttributeMap) UpdateAddIntValue(key string, value int64) {
	if v, ok := attributes.values[key]; ok {
		v.(*intValue).value = value
	} else {
		attributes.AddIntValue(key, value)
	}
}

func (attributes *AttributeMap) GetBoolValue(key string) bool {
	value := attributes.values[key]
	if x, ok := value.(*boolValue); ok {
		return x.value
	}
	return false
}

func (attributes *AttributeMap) AddBoolValue(key string, value bool) {
	attributes.values[key] = &boolValue{
		value: value,
	}
}

func (attributes *AttributeMap) UpdateAddBoolValue(key string, value bool) {
	if v, ok := attributes.values[key]; ok {
		v.(*boolValue).value = value
	} else {
		attributes.AddBoolValue(key, value)
	}
}

func (attributes *AttributeMap) RemoveAttribute(key string) {
	delete(attributes.values, key)
}

func (attributes *AttributeMap) ClearAttributes() {
	attributes.values = make(map[string]AttributeValue)
}

func (attributes *AttributeMap) ToStringMap() map[string]string {
	stringMap := make(map[string]string)
	for k, v := range attributes.values {
		stringMap[k] = v.ToString()
	}
	return stringMap
}

func (attributes *AttributeMap) GetValues() map[string]AttributeValue {
	if attributes != nil {
		return attributes.values
	}
	return nil
}

// ResetValues sets the default value for all elements. Used for implementing sync.Pool.
func (attributes *AttributeMap) ResetValues() {
	for _, v := range attributes.values {
		v.Reset()
	}
}

func (attributes *AttributeMap) String() string {
	json, _ := json.Marshal(attributes.ToStringMap())
	return string(json)
}

type AttributeValue interface {
	ToString() string
	Reset()
}

type stringValue struct {
	value string
}

func (v *stringValue) ToString() string {
	return v.value
}

func (v *stringValue) Reset() {
	v.value = ""
}

type intValue struct {
	value int64
}

func (v *intValue) ToString() string {
	return strconv.FormatInt(v.value, 10)
}

func (v *intValue) Reset() {
	v.value = 0
}

type boolValue struct {
	value bool
}

func (v *boolValue) ToString() string {
	return strconv.FormatBool(v.value)
}

func (v *boolValue) Reset() {
	v.value = false
}
