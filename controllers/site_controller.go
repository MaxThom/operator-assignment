/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tmv1 "github.com/maxthom/rocketlab-controller/api/v1"
)

// SiteReconciler reconciles a Site object
type SiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type SiteConfig struct {
	ctx  context.Context
	site *tmv1.Site
	req  ctrl.Request
	log  logr.Logger
}

// +kubebuilder:rbac:groups=tm.rocketlab.global,resources=sites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tm.rocketlab.global,resources=sites/status,verbs=get;update;patch

func (r *SiteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.Log.Info("Event " + req.Name + " on " + req.Namespace)

	// Get site of request
	var site tmv1.Site
	if err := r.Get(ctx, req.NamespacedName, &site); err != nil {
		return ResolveIfNotFound(err)
	}
	config := SiteConfig{ctx: ctx, site: &site, log: r.Log, req: req}

	if site.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted.
		// Registering our finalizer.
		if err := r.registerFinalizer(config); err != nil {
			return ctrl.Result{}, err
		}

		// Bootstrap site.
		if err := r.bootstrapSite(config); err != nil {
			return ctrl.Result{}, err
		}
	} else if containsString(site.ObjectMeta.Finalizers, siteFinalizerName) {
		// Object being deleted.
		// Takedown site.
		if err := r.takedownSite(config); err != nil {
			return ctrl.Result{}, err
		}

		// unregister our finalizer from the list.
		if err := r.unregisterFinalizer(config); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmv1.Site{}).
		Complete(r)
}

func (r *SiteReconciler) registerFinalizer(config SiteConfig) error {
	controllerutil.AddFinalizer(config.site, siteFinalizerName)
	if err := r.Update(config.ctx, config.site); err != nil {
		return err
	}

	return nil
}

func (r *SiteReconciler) unregisterFinalizer(config SiteConfig) error {
	controllerutil.RemoveFinalizer(config.site, siteFinalizerName)
	if err := r.Update(context.Background(), config.site); err != nil {
		return err
	}

	return nil
}

func (r *SiteReconciler) bootstrapSite(config SiteConfig) error {
	// Get list of tmsource with site name equal to this site
	tmSources, err := r.getTmSourcesWithSite(config)
	if err != nil {
		r.Log.Info("unable to fetch TmSources")
		return err
	}

	if config.site.Spec.Enabled {
		r.Log.Info("Site is enabled, activating tmsources...")
		// Activate all pods
		for _, tm := range tmSources {
			r.activateTmSourcePod(getPodObject(tm))
		}

	} else {
		r.Log.Info("Site is desabled, deactivating tmsources...")
		// Deactivate all pods
		for _, tm := range tmSources {
			r.deactivateTmSourcePod(getPodObject(tm))
		}
	}

	return nil
}

func (r *SiteReconciler) takedownSite(config SiteConfig) error {
	// Delete the tmsources linked to the site
	// Get list of tmsource with site name equal to this site
	tmSources, err := r.getTmSourcesWithSite(config)
	if err != nil {
		r.Log.Info("unable to fetch TmSources")
		return err
	}

	for _, tm := range tmSources {
		r.Log.Info("Deleting TmSource " + tm.Name)
		if err := r.Delete(config.ctx, &tm); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

func (r *SiteReconciler) getTmSourcesWithSite(config SiteConfig) ([]tmv1.TmSource, error) {
	// Get list of tmsource with site name equal to this site
	var tmSources tmv1.TmSourceList
	r.Log.Info("Fetching list of tmsources for site")
	err := r.List(config.ctx, &tmSources, &client.ListOptions{})
	if err != nil {
		r.Log.Info("unable to fetch TmSources")
		return nil, err
	}

	var sources []tmv1.TmSource
	for _, tm := range tmSources.Items {
		if tm.Spec.Site == config.site.Name {
			sources = append(sources, tm)
		}
	}

	return sources, nil
}

func (r *SiteReconciler) activateTmSourcePod(pod *v1.Pod) error {
	if !doesPodExist(r.Client, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}) {
		r.Log.Info("Creating pod " + pod.Name + "...")
		if err := createPod(r.Client, pod); err != nil {
			r.Log.Info("Could not create pod " + pod.Name + ".")
			return err
		}
	} else {
		r.Log.Info("Pod " + pod.Name + " already up.")
	}
	return nil
}

func (r *SiteReconciler) deactivateTmSourcePod(pod *v1.Pod) error {
	if doesPodExist(r.Client, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}) {
		r.Log.Info("Deleting pod " + pod.Name + "...")
		if err := deletePod(r.Client, pod); err != nil {
			r.Log.Info("Could not delete pod " + pod.Name + ".")
			return err
		}
	} else {
		r.Log.Info("Pod " + pod.Name + " already down.")
	}
	return nil
}
