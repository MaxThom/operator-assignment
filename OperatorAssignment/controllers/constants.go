package controllers

const (
	podTerminationWaitTimeSec = 20
	tmSourceFinalizerName     = "tmsource.finalizers.rocket.global"
	siteFinalizerName         = "site.finalizers.rocket.global"

	tmLabelSiteKey  = "site"
	tmLabelAppKey   = "app"
	tmLabelAppValue = "rocket-source-pod"
	tmNamePrefix    = "rocket-source-pod-"

	tmContainerName         = "rocket-source"
	tmContainerPath         = "maxthom/rocket-source:latest"
	tmContainerEnvMetricKey = "METRIC_NAME"
	tmContainerEnvNatKey    = "NATS_SERVICE_PORT"
	tmContainerEnvNatValue  = "nats-server-service.default.svc.cluster.local:4222"
)
