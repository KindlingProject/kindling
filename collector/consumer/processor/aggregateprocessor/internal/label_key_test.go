package internal

import (
	"github.com/Kindling-project/kindling/collector/model"
	"reflect"
	"testing"
)

func TestLabelSelectors_GetLabelKeys(t *testing.T) {
	type fields struct {
		selectors []LabelSelector
	}
	type args struct {
		labels *model.AttributeMap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LabelKeys
	}{
		{
			fields: fields{selectors: []LabelSelector{
				{Name: "stringKey", VType: StringType},
				{Name: "booleanKey", VType: BooleanType},
				{Name: "intKey", VType: IntType},
			}},
			args: args{labels: createAttributeMap()},
			want: wantLabelKeys(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LabelSelectors{
				selectors: tt.fields.selectors,
			}
			if got := s.GetLabelKeys(tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabelKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLabelKeys_GetLabels(t *testing.T) {
	type fields struct {
		keys [34]LabelKey
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.AttributeMap
	}{
		{
			fields: fields{wantLabelKeys().keys},
			want:   createAttributeMap(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &LabelKeys{
				keys: tt.fields.keys,
			}
			if got := k.GetLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createAttributeMap() *model.AttributeMap {
	ret := model.NewAttributeMap()
	ret.AddIntValue("intKey", 100)
	ret.AddStringValue("stringKey", "stringValue")
	ret.AddBoolValue("booleanKey", true)
	return ret
}

func wantLabelKeys() *LabelKeys {
	ret := &LabelKeys{}
	ret.keys[0] = LabelKey{
		Name:  "stringKey",
		Value: "stringValue",
		VType: StringType,
	}
	ret.keys[1] = LabelKey{
		Name:  "booleanKey",
		Value: "true",
		VType: BooleanType,
	}
	ret.keys[2] = LabelKey{
		Name:  "intKey",
		Value: "100",
		VType: IntType,
	}
	return ret
}
