package main

import (
	"fmt"
	"time"

	"k8s.io/klog"
)

func main() {
	metricsClient,err := BuildMetricsClient()
	if err != nil {
		klog.Fatalf("Build metrics client failed: %v", err)
	}
	kubeClient, err := BuildInClusterClientSet()
	if err != nil {
		klog.Fatalf("Build kube client failed: %v", err)
	}
	vpaController := VPAController{
		metricsClient: metricsClient,
		kubeClient: kubeClient,
	}

	for {
		err := vpaController.VerticalHorizonScaleForOnePod("test", "default", 2, 0.8)
		if err != nil {
			klog.Fatalf("vpa failed: %v", err)
			break
		}
		time.Sleep(15 * time.Second)
	}
	fmt.Println("Hello world!")
}