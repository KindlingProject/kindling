package service

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestMetaDataWrapper_ListAndWatch(t *testing.T) {
	mp, err := NewMetaDataWrapper(&Config{
		KubeAuthType:  "kubeConfig",
		KubeConfigDir: "/root/.kube/config",
	})
	if err != nil {
		fmt.Println(err)
	}
	http.HandleFunc("/listAndWatch", mp.ListAndWatch)
	// Add Func

	log.Fatal(http.ListenAndServe(":8081", nil))
}
