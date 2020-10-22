package main

import (
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/custom_metrics"
	"k8s.io/metrics/pkg/client/external_metrics"
)

// BuildMetrics builds the metrics client
func BuildMetricsClient() (metrics.MetricsClient, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to build cluster config: %v", err)
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Failed to build client: %v", err)
	}
	cachedClient := cacheddiscovery.NewMemCacheClient(client.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedClient)
	apiVersionsGetter := custom_metrics.NewAvailableAPIsGetter(client.Discovery())
	metricsClient := metrics.NewRESTMetricsClient(
		resourceclient.NewForConfigOrDie(cfg),
		custom_metrics.NewForConfig(cfg, restMapper, apiVersionsGetter),
		external_metrics.NewForConfigOrDie(cfg),
	)
	return metricsClient, nil
}

func BuildInClusterClientSet() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to build cluster config: %v", err)
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}