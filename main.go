package main

import (
	log "github.com/sirupsen/logrus"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"syscall"
)

const curlAnnotation = "x-k8s.io/curl-me-that"

// curlConfigMap will fetch the pages referenced to by the curlAnnotation and
// update the ConfigMap data with them.
func curlConfigMap(configMap ConfigMap) {
	logger := log.WithFields(log.Fields{
		"name":      configMap.Name,
		"namespace": configMap.Namespace,
	})

	urls, ok := configMap.Annotations[curlAnnotation]
	if !ok {
		logger.Info("Skipping configmap without annotation")
		return
	}

	fetcher, err := PageFetcherFromString(urls)
	if err != nil {
		logger.WithError(err).Warning("Cannot parse URLs")
		configMap.RecordWarning("Can't parse URL: %v", err)
		return
	}

	// Do not fetch what the ConfigMap already has, as an optimization,
	// but also to prevent an infinite loop of constantly refreshing the
	// data in the configmap.
	fetcher.Exclude(configMap.Data)

	data, err := fetcher.Fetch()
	if err != nil {
		logger.WithError(err).Warning("Cannot fetch URLs")
		configMap.RecordWarning("Can't fetch URL: %v", err)
		// Do not return here, set the data on a best-effort basis.
	}

	if len(data) == 0 {
		logger.Info("Leaving configmap already processed")
		return
	}

	err = configMap.Push(data)
	if err != nil {
		logger.WithError(err).Error("Cannot add data")
		return
	}

	logger.WithField("data", data).Info("Curled data into ConfigMap")
}

func main() {
	// If $KUBECONFIG is defined, use its configuration.
	// Otherwise fall back to in-cluster config.
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		log.WithError(err).Fatal("Cannot load config")
	}

	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("Creating kubernetes clientset")
	}

	manager := NewConfigMapManager(clientset)
	configmaps, err := manager.StartWatching()
	if err != nil {
		log.WithError(err).Fatal("Cannot watch ConfigMaps")
	}

	// Listen to SIGINT or SIGTERM to cleanly exit after
	// either Ctrl-C or Kubernetes Pod termination.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-stop
		log.WithField("signal", sig).Info("Received signal for termination")
		manager.StopWatching()
	}()

	for configmap := range configmaps {
		go curlConfigMap(configmap)
	}
}
