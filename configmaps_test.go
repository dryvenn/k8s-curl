package main

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func testChanEmpty(t *testing.T, c chan interface{}) {
	// Sleep a bit for propagation to happen
	time.Sleep(100 * time.Millisecond)
	if len(c) != 0 {
		t.Error("chan not empty")
	}
}

func TestConfigMapWatching(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	configMapClient := clientset.CoreV1().ConfigMaps("")

	manager := NewConfigMapManager(clientset)
	notifChan, err := manager.StartWatching()
	if err != nil {
		t.Fatal(err)
	}

	testNoNotif := func() {
		time.Sleep(100 * time.Millisecond)
		if len(notifChan) != 0 {
			t.Error("configmap chan not empty")
		}
	}

	// create 1 configmap
	testConfigMap := core_v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test",
		},
	}
	configMapClient.Create(&testConfigMap)

	// receive one configmap
	notif := <-notifChan
	if notif.Name != testConfigMap.Name {
		t.Fatal("configmap name mismatch")
	}
	testNoNotif()

	// change configmap manually
	testConfigMap.Data = map[string]string{"key": "val"}
	configMapClient.Update(&testConfigMap)

	// receive one configmap
	notif = <-notifChan
	if notif.Name != testConfigMap.Name {
		t.Fatal("configmap name mismatch")
	}
	if notif.Data["key"] != testConfigMap.Data["key"] {
		t.Fatal("configmap val mismatch")
	}
	testNoNotif()

	// change configmap by update method
	err = manager.UpdateData(&testConfigMap, map[string]string{"another": "couple"})
	if err != nil {
		t.Fatal(err)
	}
	_ = <-notifChan
	testNoNotif()

	// delete configmap
	configMapClient.Delete(testConfigMap.Name, &meta_v1.DeleteOptions{})
	testNoNotif()

	manager.StopWatching()
	testNoNotif()
	_, ok := <-notifChan
	if ok {
		t.Error("configmap chan not closed")
	}
}
