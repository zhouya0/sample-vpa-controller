package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/core/helper/qos"
	"k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

type VPAController struct {
	metricsClient metrics.MetricsClient
	podLister corev1.PodLister
	kubeClient *kubernetes.Clientset
}

func (v *VPAController) VerticalHorizonScaleForOnePod(podName string, podNamespace string , scaleRate int) error {
	pod,err := v.kubeClient.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !podCanVPA(pod) {
		return fmt.Errorf("pod can't do vertical scaling")
	}
}

func podCanVPA(obj runtime.Object) bool {
	pod := obj.(*core.Pod)
	qosClass := qos.GetPodQOS(pod)
	// for now, this vpa controller only support pod with qos class guaranteed
	if qosClass != core.PodQOSGuaranteed {
		return false
	}
	return true
}

func calculatePodResourceLimit(pod *v1.Pod, resource v1.ResourceName) int64 {
	var podLimit int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := getLimitForResource(resource, &container.Resources.Requests)
		podLimit = value
	}
	return podLimit
}

func getLimitForResource(resource v1.ResourceName, limits *v1.ResourceList) int64 {
	switch resource {
	case v1.ResourceCPU:
		return limits.Cpu().MilliValue()
	case v1.ResourceMemory:
		return limits.Memory().Value()
	}
	return 0
}
