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
	"time"

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

// TmSourceReconciler reconciles a TmSource object
type TmSourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type TmSourceConfig struct {
	ctx      context.Context
	tmsource *tmv1.TmSource
	pod      *v1.Pod
	req      ctrl.Request
	log      logr.Logger
}

// +kubebuilder:rbac:groups=tm.rocketlab.global,resources=tmsources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tm.rocketlab.global,resources=tmsources/status,verbs=get;update;patch

func (r *TmSourceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.Log.Info("Event " + req.Name + " on " + req.Namespace)

	// Get current tmsource
	var tmsource tmv1.TmSource
	if err := r.Get(ctx, req.NamespacedName, &tmsource); err != nil {
		return ResolveIfNotFound(err)
	}
	config := TmSourceConfig{ctx: ctx, tmsource: &tmsource, pod: getPodObject(tmsource), log: r.Log, req: req}

	if tmsource.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted.
		// Registering our finalizer.
		if err := r.registerFinalizer(config); err != nil {
			return ctrl.Result{}, err
		}

		// Bootstrap tmsource pod.
		if err := r.bootstrapTmSourcePod(config); err != nil {
			return ctrl.Result{}, err
		}
	} else if containsString(tmsource.ObjectMeta.Finalizers, tmSourceFinalizerName) {
		// Object being deleted.
		// Takedown tmsource pod.
		if err := r.takedownTmSourcePod(config); err != nil {
			return ctrl.Result{}, err
		}

		// unregister our finalizer from the list.
		if err := r.unregisterFinalizer(config); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *TmSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmv1.TmSource{}).
		Complete(r)
}

func (r *TmSourceReconciler) registerFinalizer(config TmSourceConfig) error {
	controllerutil.AddFinalizer(config.tmsource, tmSourceFinalizerName)
	if err := r.Update(config.ctx, config.tmsource); err != nil {
		return err
	}

	return nil
}

func (r *TmSourceReconciler) unregisterFinalizer(config TmSourceConfig) error {
	controllerutil.RemoveFinalizer(config.tmsource, tmSourceFinalizerName)
	if err := r.Update(context.Background(), config.tmsource); err != nil {
		return err
	}

	return nil
}

func (r *TmSourceReconciler) bootstrapTmSourcePod(config TmSourceConfig) error {

	// Get site associated with this source
	site, err := r.getSourceSite(config)
	if err != nil {
		r.Log.Error(err, "unable to get site")
		return err
	}

	// Get pod associated with this source
	podInstance, err := r.getTmSourcePod(config)
	if err != nil {
		return err
	}
	if podInstance == nil {
		r.Log.Info("Pod of TmSource is non-existent.")
	} else {
		r.Log.Info("Pod of TmSource is " + podInstance.Name + ".")
	}

	// We still create the source even if there is no site linked
	if site == nil || site.Spec.Enabled {
		r.Log.Info("Site is enabled.")
		// In case metrics are different, delete and recreate
		if podInstance != nil {
			r.Log.Info("Deleting old pod...")
			if err := r.deleteTmSourcePod(config, true); err != nil {
				r.Log.Info("Could not delete pod " + config.pod.Name + ".")
			}
		}

		// Create pod
		r.Log.Info("Creating new pod...")
		if err := createPod(r.Client, config.pod); err != nil {
			r.Log.Error(err, "Could not create pod.")
		}

		r.Log.Info("TmSource pod is operationnal !")

		return nil
	} else if !site.Spec.Enabled {
		r.Log.Info("Site is disabled.")
		// Check if exist, if so delete
		if podInstance != nil {
			r.Log.Info("Deleting old pod...")
			if err := r.deleteTmSourcePod(config, false); err != nil {
				r.Log.Info("Could not delete pod " + config.pod.Name + ".")
			}
			r.Log.Info("TmSource pod is deleted !")
		}
	}

	return nil
}

func (r *TmSourceReconciler) takedownTmSourcePod(config TmSourceConfig) error {
	r.Log.Info("Deleting " + config.pod.Name + " pod.")

	if err := r.deleteTmSourcePod(config, false); err != nil {
		r.Log.Info("Could not delete pod " + config.pod.Name + ".")
		return err
	}

	r.Log.Info("TmSource pod is deleted !")
	return nil
}

func (r *TmSourceReconciler) getSourceSite(config TmSourceConfig) (*tmv1.Site, error) {
	var site tmv1.Site
	if err := r.Get(config.ctx, types.NamespacedName{Name: config.tmsource.Spec.Site, Namespace: config.tmsource.Namespace}, &site); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("No site linked to the source")
			return nil, nil
		}
		return nil, err
	}

	r.Log.Info("Site of TmSource is " + site.Name + ".")
	return &site, nil
}

func (r *TmSourceReconciler) getTmSourcePod(config TmSourceConfig) (*v1.Pod, error) {
	var pod v1.Pod
	loc := types.NamespacedName{
		Name:      config.pod.Name,
		Namespace: config.tmsource.Namespace,
	}

	if err := r.Get(config.ctx, loc, &pod); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &pod, nil
}

func (r *TmSourceReconciler) deleteTmSourcePod(config TmSourceConfig, waitForTermination bool) error {
	if err := deletePod(r.Client, config.pod); err != nil {
		r.Log.Info("Could not delete pod " + config.pod.Name + ".")
		return err
	}

	if waitForTermination {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			i++

			pod, err := r.getTmSourcePod(config)
			if err != nil {
				return err
			}
			if pod == nil {
				return nil
			}

			if i > podTerminationWaitTimeSec {
				return nil
			}
		}
	}

	return nil
}
