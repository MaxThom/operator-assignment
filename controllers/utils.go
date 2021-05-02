package controllers

import (
	"context"

	tmv1 "github.com/maxthom/rocketlab-controller/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RequeueError struct{}

func (m *RequeueError) Error() string {
	return "requeue"
}

func doesPodExist(c client.Client, loc types.NamespacedName) bool {
	return c.Get(context.Background(), loc, &v1.Pod{}) == nil
}

func createPod(c client.Client, pod *v1.Pod) error {
	return c.Create(context.Background(), pod)
}

func deletePod(c client.Client, pod *v1.Pod) error {
	if err := c.Delete(context.Background(), pod); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func getPodObject(tmsource tmv1.TmSource) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tmNamePrefix + tmsource.Name,
			Namespace: tmsource.Namespace,
			Labels: map[string]string{
				tmLabelAppKey: tmLabelAppValue,
				//tmLabelSiteKey: tmsource.Spec.Site,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            tmContainerName,
					Image:           tmContainerPath,
					ImagePullPolicy: v1.PullAlways,
					Env: []v1.EnvVar{
						{
							Name:  tmContainerEnvNatKey,
							Value: tmContainerEnvNatValue,
						},
						{
							Name:  tmContainerEnvMetricKey,
							Value: tmsource.Spec.MetricName,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{},
	}
}

func isPodEnvDifferent(a *v1.Pod, b *v1.Pod) bool {
	for i, env := range a.Spec.Containers[0].Env {
		if env.Value != b.Spec.Containers[0].Env[i].Value {
			return true
		}
	}

	return false
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ResolveIfNotFound(err error) (ctrl.Result, error) {
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}
