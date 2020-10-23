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
	if threshold < 0 || threshold > 1 {
		return fmt.Errorf("threshold should be a value between 0 and 1, but got: %v", threshold)
	}
	if scaleRate < 1 || scaleRate > 10{
		return fmt.Errorf("scaleRate too small or too large: %v", scaleRate)
	}
	pod,err := v.kubeClient.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	//if !podCanVPA(pod) {
	//	return fmt.Errorf("pod can't do vertical scaling")
	//}
	specCPU := calculatePodResourceLimit(pod, v1.ResourceCPU)
	specMemory := calculatePodResourceLimit(pod, v1.ResourceMemory)
	currentCPUs, _, err := v.metricsClient.GetResourceMetric(v1.ResourceCPU, podNamespace, labels.Nothing())
	if err != nil {
		return fmt.Errorf("metrics client get cpu failed: %v", err)
	}
	if _, ok := currentCPUs[podName]; !ok {
		return fmt.Errorf("pod %v not found in metric client", podName)
	}
	currentCPU := currentCPUs[podName]
	currentMemorys, _, err := v.metricsClient.GetResourceMetric(v1.ResourceMemory, podNamespace, labels.Nothing())
	if err != nil {
		return fmt.Errorf("metrics cleint get memory failed: %v", err)
	}
	if _, ok := currentMemorys[podName]; !ok {
		return fmt.Errorf("pod %v not found in metric client", podName)
	}
	currentMemory := currentMemorys[podName]
	klog.Infof("spec CPU is %v\n", specCPU)
	klog.Infof("spec Memory is %v\n", specMemory)
	klog.Infof("current CPU is %v \n", currentCPU)
	klog.Infof("current Memory is %v\n", currentMemory)

	// For this demo, we only do cpu, memory is the same logic
	newCPUValue, shouldScale := computeNewCPUValue(float32(specCPU), float32(currentCPU.Value), scaleRate, threshold)
	if shouldScale {
		newPod := generatePodResourceRequestLimit(pod, v1.ResourceCPU, newCPUValue)
		_, err := v.kubeClient.CoreV1().Pods(podNamespace).Update(context.TODO(), newPod, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func computeNewCPUValue(spec float32, current float32, scaleRate float32, threshold float32) (int64,bool) {
	thresholdCPU := spec * threshold
	if thresholdCPU < current {
		return int64(spec * scaleRate), true
	}
	return 0, false
}

// panic: interface conversion: runtime.Object is *v1.Pod, not *core.Pod
// Don't use this logic for now, it will panic.
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

func generatePodResourceRequestLimit(pod *v1.Pod, resource v1.ResourceName, newValue int64) *v1.Pod {
	klog.Infof("Pod resource before: %v", pod.Spec.Containers[0].Resources)
	switch resource {
	case v1.ResourceCPU:
		cpuResource := pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU]
		cpuResource.SetMilli(newValue)
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuResource
		cpuLimit := pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU]
		cpuLimit.SetMilli(newValue)
		pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimit
	}
	klog.Infof("Pod resource after: %v", pod.Spec.Containers[0].Resources)
	return pod
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
