package main

import (
	"context"
	"fmt"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/core/helper/qos"
	"k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

type VPAController struct {
	metricsClient metrics.MetricsClient
	kubeClient *kubernetes.Clientset
}

func (v *VPAController) VerticalHorizonScaleForOnePod(podName string, podNamespace string, scaleRate float32, threshold float32) error {
	pod,err := v.kubeClient.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !podCanVPA(pod) {
		return fmt.Errorf("pod can't do vertical scaling")
	}
	specCPU := calculatePodResourceLimit(pod, v1.ResourceCPU)
	specMemory := calculatePodResourceLimit(pod, v1.ResourceMemory)
	currentCPUs, _, err := v.metricsClient.GetResourceMetric(v1.ResourceCPU, podNamespace, labels.Nothing())
	if err != nil {
		return fmt.Errorf("metrics client get cpu failed")
	}
	if _, ok := currentCPUs[podName]; !ok {
		return fmt.Errorf("pod %v not found in metric client", podName)
	}
	currentCPU := currentCPUs[podName]
	currentMemorys, _, err := v.metricsClient.GetResourceMetric(v1.ResourceMemory, podNamespace, labels.Nothing())
	if err != nil {
		return fmt.Errorf("metrics cleint get memory failed")
	}
	if _, ok := currentMemorys[podName]; !ok {
		return fmt.Errorf("pod %v not found in metric client", podName)
	}
	currentMemory := currentMemorys[podName]
	klog.Infof("spec CPU is %v\n", specCPU)
	klog.Infof("spec Memory is %v\n", specMemory)
	klog.Infof("current CPU is %v \n", currentCPU)
	klog.Infof("current Memory is %v\n", currentMemory)
	return nil
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
