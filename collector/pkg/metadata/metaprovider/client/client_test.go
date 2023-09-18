package metadataclient

import (
	"fmt"
	"testing"
)

func TestClient_ListAndWatch(t *testing.T) {
	for i := 0; i < 1000; i++ {
		go func() {
			cli := NewMetaDataWrapperClient("http://localhost:9504", true)
			err := cli.ListAndWatch(nil)
			if err != nil {
				fmt.Println(err)
			}
		}()
		fmt.Printf("init: %d\n", i)
	}

	select {}
}

// func TestClient_ListAndWatch(t *testing.T) {
// 	cli := NewMetaDataWrapperClient("http://10.10.101.69:59504", true)
// 	log.Fatal(cli.ListAndWatch())
// }
