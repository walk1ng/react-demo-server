package k8s

import (
	"context"
	"fmt"
	"os/user"
	"path"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var cfg *rest.Config
var client *kubernetes.Clientset
var err error

type PodInfo struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Node      string `json:"node,omitempty"`
	UID       string `json:"uid,omitempty"`
}

type NsInfo struct {
	Name string `json:"name,omitempty"`
	UID  string `json:"uid,omitempty"`
}

func init() {
	// k8s
	cfg, err = getK8sConfig()
	client, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("成功获取k8s client")
}

func GetPodsByNS(namespace string) ([]PodInfo, error) {
	if namespace == "all" {
		namespace = ""
	}

	podClient := client.CoreV1().Pods(namespace)

	pl, err := podClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var pods []corev1.Pod
	pods = append(pods, pl.Items...)

	var podInfo []PodInfo
	for _, pod := range pods {
		podInfo = append(podInfo, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Phase:     string(pod.Status.Phase),
			Node:      pod.Spec.NodeName,
			UID:       string(pod.UID),
		})

	}
	return podInfo, nil
}

func GetNs() ([]NsInfo, error) {
	nsClient := client.CoreV1().Namespaces()
	nl, err := nsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nsInfo []NsInfo
	for _, ns := range nl.Items {
		nsInfo = append(nsInfo, NsInfo{
			Name: ns.Name,
			UID:  string(ns.UID),
		})
	}
	return nsInfo, nil
}

func getK8sConfig() (*rest.Config, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	cfgPath := path.Join(u.HomeDir, ".kube", "config")

	cfg, err = clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
