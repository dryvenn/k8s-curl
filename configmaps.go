package main

import (
	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

// ConfigMapManager is a simple wrapper around the Kubernetes API
// that provides app-specific helpers from on base primitives.
type ConfigMapManager struct {
	clientset   k8s.Interface
	watchHandle watch.Interface
	recorder    record.EventRecorder
}

type ConfigMap struct {
	*core_v1.ConfigMap
	manager *ConfigMapManager
}

// NewConfigMapManager creates a new ConfigMapManager
func NewConfigMapManager(clientset k8s.Interface) *ConfigMapManager {
	// Make a recorder for events
	// See https://git.io/JfNVp
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&typed_core_v1.EventSinkImpl{
		Interface: clientset.CoreV1().Events(""),
	})
	recorder := broadcaster.NewRecorder(scheme.Scheme, core_v1.EventSource{Component: "k8s-curl"})
	return &ConfigMapManager{
		clientset: clientset,
		recorder:  recorder,
	}
}

// StartWatching returns a channel in which ConfigMap objects will be sent
// whenever a ConfigMap is added or changed in the cluster.
// It will ignore events from errors and deletions.
// Subsequent calls to this method will cancel previous ones.
func (m *ConfigMapManager) StartWatching() (<-chan ConfigMap, error) {
	// Make sure Watch wasn't already ongoing.
	m.StopWatching()

	// TODO: Should this be used for a controller? Have a look at
	// https://godoc.org/k8s.io/client-go/tools/cache
	handle, err := m.clientset.CoreV1().ConfigMaps("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	m.watchHandle = handle

	// TODO: determine the correct buffering here
	configMaps := make(chan ConfigMap, 100)

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
func (m *ConfigMapManager) processEvents(events <-chan watch.Event, configMaps chan<- ConfigMap) {
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
			configMaps <- ConfigMap{ConfigMap: cm, manager: m}
		}
	}
	close(configMaps)
}

// Push adds the given data to the Data field of the configmap and sends
// the update to Kubernetes.
// Conflicting keys are overwritten.
// Note that in case of error, the ConfigMap struct will still be modified even
// if this change is not reflected in Kubernetes!
func (cm ConfigMap) Push(data map[string]string) error {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	for k, v := range data {
		cm.Data[k] = v
	}

	var err error
	for i := 0; i < 3; i++ {
		_, err := cm.manager.clientset.CoreV1().ConfigMaps(cm.Namespace).Update(cm.ConfigMap)
		if err == nil {
			break
		}
	}
	return err
}

// RecordWarning records an event on the configmap resource of type Warning
func (cm *ConfigMap) RecordWarning(fmt string, args ...interface{}) {
	cm.manager.recorder.Eventf(cm.ConfigMap, core_v1.EventTypeWarning, "k8s-url", fmt, args...)
}
