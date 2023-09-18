package metadataclient

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/api"
)

type SetPreprocessingMetaDataCache func(cache *kubernetes.K8sMetaDataCache, nodeMap *kubernetes.NodeMap, serviceMap *kubernetes.ServiceMap, rsMap *kubernetes.ReplicaSetMap)

type Client struct {
	// tls wrap
	cli      http.Client
	endpoint string
	debug    bool
}

func NewMetaDataWrapperClient(endpoint string, debug bool) *Client {
	return &Client{
		cli:      *createHTTPClient(),
		debug:    debug,
		endpoint: endpoint + "/listAndWatch",
	}
}

func (c *Client) ListAndWatch(setup SetPreprocessingMetaDataCache) error {
	// handler cache.ResourceEventHandler,
	resp, err := c.cli.Get(c.endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 500 {
		return errors.New("Provider Server is not ready yet,please wait")
	}
	reader := bufio.NewReaderSize(resp.Body, 1024*32)
	b, _ := reader.ReadBytes('\n')
	// var listData api.MetaData
	// json.Unmarshal(b, &listData)
	// cache := kubernetes.K8sMetaDataCache{}
	listVO := api.ListVO{}
	err = json.Unmarshal(b, &listVO)
	if err != nil {
		// 本次连接失败，等待重试
		return err
	}
	if c.debug {
		formatCache, _ := json.MarshalIndent(listVO.Cache, "", "\t")
		log.Printf("K8sCache Init:%s\n", string(formatCache))
	}

	setup(listVO.Cache, listVO.GlobalNodeInfo, listVO.GlobalServiceInfo, listVO.GlobalRsInfo)

	for {
		b, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("remote server send unexpected EOF,shutting down,err:%v", err)
				break
			} else {
				log.Printf("receive unexpected error durning watch ,err:%v", err)
				break
			}
		}
		var resp api.MetaDataVO
		err = json.Unmarshal(b, &resp)
		if err != nil {
			log.Printf("parse response failed ,err:%v", err)
			continue
		}
		c.apply(&resp)
		if kubernetes.ReWatch {
			kubernetes.ReWatch = false
			break
		}
	}
	return nil
}

func (c *Client) apply(resp *api.MetaDataVO) {
	switch resp.Type {
	case "pod":
		podUnwrapperHander.Apply(resp)
	case "service":
		serviceUnwrapperHander.Apply(resp)
	case "rs":
		relicaSetUnwrapperHander.Apply(resp)
	case "node":
		nodeUnwrapperHander.Apply(resp)
	default:
		// TODO Detail

	}
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
	return client
}
