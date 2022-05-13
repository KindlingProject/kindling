// Copyright 2020 OpenTelemetry Authors
// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/k8sattributesprocessor/internal/kube/client.go
package kubernetes

import (
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
)

var (
	podDeleteQueueMut sync.Mutex
	podDeleteQueue    []deleteRequest
)

type deleteRequest struct {
	pod *corev1.Pod
	ts  time.Time
}

// deleteLoop deletes pods from cache periodically.
func podDeleteLoop(interval time.Duration, gracePeriod time.Duration, stopCh chan struct{}) {
	// This loop runs after N seconds and deletes pods from cache.
	// It iterates over the delete queue and deletes all that aren't
	// in the grace period anymore.
	for {
		select {
		case <-time.After(interval):
			var cutoff int
			now := time.Now()
			podDeleteQueueMut.Lock()
			for i, d := range podDeleteQueue {
				if d.ts.Add(gracePeriod).After(now) {
					break
				}
				cutoff = i + 1
			}
			toDelete := podDeleteQueue[:cutoff]
			podDeleteQueue = podDeleteQueue[cutoff:]
			podDeleteQueueMut.Unlock()
			for _, d := range toDelete {
				deletePod(d.pod)
			}

		case <-stopCh:
			return
		}
	}
}

func deletePod(pod *corev1.Pod) {
	for i := 0; i < len(pod.Status.ContainerStatuses); i++ {
		containerId := pod.Status.ContainerStatuses[i].ContainerID
		realContainerId := TruncateContainerId(containerId)
		if realContainerId == "" {
			continue
		}
		MetaDataCache.DeleteByContainerId(realContainerId)
	}

	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			// Assume that PodIP:Port can't be reused in a few seconds
			MetaDataCache.DeleteContainerByIpPort(pod.Status.PodIP, uint32(port.ContainerPort))
			// If hostPort is specified, add the container using HostIP and HostPort
			if port.HostPort != 0 {
				MetaDataCache.DeleteContainerByHostIpPort(pod.Status.HostIP, uint32(port.HostPort))
			}
		}
	}
}
