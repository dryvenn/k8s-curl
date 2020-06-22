package main

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
)

func TestCurlConfigMap(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	configMap := ConfigMap{
		ConfigMap: &core_v1.ConfigMap{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					curlAnnotation: "datanet=data.net",
				},
			},
		},
		manager: NewConfigMapManager(clientset),
	}
	clientset.CoreV1().ConfigMaps("").Create(configMap.ConfigMap)

	curlConfigMap(configMap)

	testConfigMap := func(cm *core_v1.ConfigMap) {
		if val, ok := cm.Data["datanet"]; !ok {
			t.Error("configMap data missing key")
		} else if !strings.Contains(val, "folks") {
			t.Error("configMap data incorret val")
		}
	}

	testConfigMap(configMap.ConfigMap)
	cm, err := clientset.CoreV1().ConfigMaps("").Get(configMap.Name, meta_v1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	testConfigMap(cm)
}
