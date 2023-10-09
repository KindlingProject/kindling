package kubernetes

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	clientSet, err := initClientSet("kubeConfig", "/Users/mr.wang/Desktop/config")
	if err != nil {
		t.Fatalf("cannot init clientSet, %s", err)
	}
	go NodeWatch(clientSet)
	go RsWatch(clientSet)
	go ServiceWatch(clientSet)
	go PodWatch(clientSet, 60*time.Second)
	time.Sleep(2 * time.Second)
	content, _ := json.Marshal(GlobalRsInfo)
	fmt.Println(string(content))
	content, _ = json.Marshal(GlobalServiceInfo)
	fmt.Println(string(content))
	content, _ = json.Marshal(GlobalPodInfo)
	fmt.Println(string(content))
	content, _ = json.Marshal(GlobalNodeInfo)
	fmt.Println(string(content))
}
