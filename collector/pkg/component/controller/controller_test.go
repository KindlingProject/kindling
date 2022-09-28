package controller

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestController(t *testing.T) {
	req := `{"operation":"start","options":{"duration":"60s","dataSize":300}}`

	reqVo := &ControlRequest{}
	json.Unmarshal([]byte(req), reqVo)

	ProfileOption := &ProfileOption{}
	json.Unmarshal(*reqVo.Options, ProfileOption)

	fmt.Printf("%+v\n", ProfileOption)
}
