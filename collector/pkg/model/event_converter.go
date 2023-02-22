package model

import "github.com/Kindling-project/kindling/collector/pkg/model/constnames"

// ConvertSendmmsg converts the single sendmmsg event to multiple ones by parsing
// the data field. The original data is composed of the loop of "size(uint32)+data".
func ConvertSendmmsg(evt *KindlingEvent) []*KindlingEvent {
	ret := make([]*KindlingEvent, 0)
	if evt.Name != constnames.SendMMsgEvent {
		return ret
	}

	data := evt.GetData()
	bytesSlice := splitDataBytes(data)
	for _, bytes := range bytesSlice {
		evtTemplate := *evt
		evtTemplate.GetUserAttribute("data").Value = bytes
		ret = append(ret, &evtTemplate)
	}
	return ret
}

func splitDataBytes(data []byte) [][]byte {
	ret := make([][]byte, 0)
	dataLength := len(data)
	for startIndex := 0; startIndex < dataLength; {
		sizeEndIndex := startIndex + 4
		if sizeEndIndex >= dataLength {
			break
		}
		sizeBytes := data[startIndex:sizeEndIndex]
		size := byteOrder.Uint32(sizeBytes)
		fragmentEnd := sizeEndIndex + int(size)
		var childData []byte
		if fragmentEnd < dataLength {
			childData = data[sizeEndIndex:fragmentEnd]
		} else {
			childData = data[sizeEndIndex:]
		}
		ret = append(ret, childData)
		startIndex = fragmentEnd
	}
	return ret
}
