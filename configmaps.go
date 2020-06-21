package main

import (
	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
)

// ConfigMapManager is a simple wrapper around the Kubernetes API
// that provides app-specific helpers from on base primitives.
type ConfigMapManager struct {
	clientset   k8s.Interface
	watchHandle watch.Interface
}

// NewConfigMapManager creates a new ConfigMapManager
func NewConfigMapManager(clientset k8s.Interface) *ConfigMapManager {
	return &ConfigMapManager{
		clientset: clientset,
	}
}

// StartWatching returns a channel in which ConfigMap objects will be sent
// whenever a ConfigMap is added or changed in the cluster.
// It will ignore events from errors and deletions.
func (m *ConfigMapManager) StartWatching() (<-chan *core_v1.ConfigMap, error) {
	m.StopWatching()

	handle, err := m.clientset.CoreV1().ConfigMaps("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	m.watchHandle = handle

	// TODO: determine the correct buffering here
	configMaps := make(chan *core_v1.ConfigMap, 100)

	go m.processEvents(handle.ResultChan(), configMaps)

	return configMaps, nil
}

// StopWatching triggers cancellation of ConfigMap watching, eventually leading
// to the closure of the channel returned by StartWatching().
func (m *ConfigMapManager) StopWatching() {
	if m.watchHandle == nil {
		return
	}
	m.watchHandle.Stop()
	m.watchHandle = nil
}

// processEvents transforms event objects into configmap objects, and filters
// events by type.
func (m *ConfigMapManager) processEvents(events <-chan watch.Event, configMaps chan<- *core_v1.ConfigMap) {
EventLoop:
	for event := range events {
		// Only keep mutations, ignore deletions and errors
		switch event.Type {
		case watch.Added, watch.Modified:
		default:
			continue EventLoop
		}
		if cm, ok := event.Object.(*core_v1.ConfigMap); !ok {
			log.WithField("obj", event.Object).Error("Received event for not a configmap")
		} else {
			configMaps <- cm
		}
	}
	close(configMaps)
}

// UpdateData updates the Data field of the configmap.
func (m *ConfigMapManager) UpdateData(configMap *core_v1.ConfigMap, data map[string]string) error {
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	for k, v := range data {
		configMap.Data[k] = v
	}

	var err error
	for i := 0; i < 3; i++ {
		_, err := m.clientset.CoreV1().ConfigMaps(configMap.Namespace).Update(configMap)
		if err == nil {
			break
		}
	}
	return err
}
