/*
Copyright 2021.

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
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/imroc/req"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var watchedRegistry = getEnv("WATCHED_REGISTRY", "harbor.mytanzu.xyz")

var prefixImageLabel = getEnv("PREFIX_IMAGE_LABEL", "kpack.")

var prefixPodLabel = getEnv("PREFIX_POD_LABEL", "tanzu-build-service")

var requestDebug = getEnv("REQUEST_DEBUG", "false") == "true"

func splitImage(image string) (domain string, remainder string, tag string) {
	i := strings.IndexRune(image, '/')
	domain, remainder = image[:i], image[i+1:]
	itag := strings.IndexRune(remainder, ':')
	remainder, tag = remainder[:itag], remainder[itag+1:]
	return
}

type Manifests struct {
	SchemaVersion string
	MediaType     string
	Config        Config
}

//Config Structure
type Config struct {
	MediaType string
	Size      string
	Digest    string
}

func queryDigest(ctx context.Context, image string) (digest string) {
	header := req.Header{
		"Accept":  "application/vnd.docker.distribution.manifest.v2+json",
		"Expires": "10ms",
	}

	req.Debug = requestDebug
	domain, repo, tag := splitImage(image)
	manifest_registry_url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", domain, repo, tag)

	log.FromContext(ctx).Info("===> queryDigest Connecting URL " + manifest_registry_url)
	r, _ := req.Get(manifest_registry_url, header)
	var result Manifests
	r.ToJSON(&result)
	//log.FromContext(ctx).Info("===> queryDigest Digest "+result.Config.Digest)
	return result.Config.Digest
}

type BlobConfig struct {
	Image  string
	User   string
	Labels map[string]string
}
type BlobResult struct {
	Architecture string
	Created      string
	Os           string
	Config       BlobConfig
}

func queryConfig(ctx context.Context, image string, digest string) (config BlobConfig) {
	header := req.Header{
		"Accept":  "application/vnd.docker.distribution.manifest.v2+json",
		"Expires": "10ms",
	}

	req.Debug = requestDebug
	domain, repo, _ := splitImage(image)
	blob_registry_url := fmt.Sprintf("https://%s/v2/%s/blobs//%s", domain, repo, digest)

	log.FromContext(ctx).Info("===> queryConfig Connecting URL" + blob_registry_url)
	r, _ := req.Get(blob_registry_url, header)
	var result BlobResult

	r.ToJSON(&result)
	//log.FromContext(ctx).Info("===> queryConfig Digest"+ result.Config.Labels)
	config = result.Config
	return
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func filterImageLabels(input map[string]string, prefix string) map[string]string {
	var candidates = make(map[string]string)
	for key, element := range input {
		//fmt.Println("FILTER prefix", prefix, "Key:", key, "=>", "Element:", element)
		if strings.HasPrefix(key, prefix) {
			//fmt.Println("Add Key:", key, "=>", "Element:", element)
			candidates[key] = element
		}
	}
	return candidates
}

func isAllTheLabelsSet(pod corev1.Pod, labels map[string]string) bool {
	if pod.Labels == nil {
		return false
	}

	for k, v := range labels {
		if val, ok := pod.Labels[k]; ok {
			//key is in the pod.Labesl
			if val == v {
				//and the value is the same
			} else {
				//but the value isn't the same
				return false
			}
		} else {
			//key not found
			return false
		}
	}
	return true
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	/*
	   Step 0: Fetch the Pod from the Kubernetes API.
	*/

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since we can get them on deleted requests.
			return ctrl.Result{}, nil
		}
		log.FromContext(ctx).Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	/*
		stop -1 : get images in the pod
	*/
	var all_candidates = make(map[string]string)
	for container := range pod.Spec.Containers {
		image := pod.Spec.Containers[container].Image
		if strings.HasPrefix(image, watchedRegistry) {
			log.FromContext(ctx).Info("===> image " + image)
			domain, repository, tag := splitImage(image)
			log.FromContext(ctx).Info("===> domain " + domain)
			log.FromContext(ctx).Info("===> repository " + repository)
			log.FromContext(ctx).Info("===> tag " + tag)
			digest := queryDigest(ctx, image)
			log.FromContext(ctx).Info("===> digest " + digest)
			config := queryConfig(ctx, image, digest)

			for k, v := range filterImageLabels(config.Labels, prefixImageLabel) {
				all_candidates[prefixPodLabel+"/"+k] = v
			}
		}
	}

	if len(all_candidates) == 0 {
		// The desired state and actual state of the Pod are the same.
		// No further action is required by the operator at this moment.
		log.FromContext(ctx).Info("no kpackLabelsFound")
		return ctrl.Result{}, nil
	}

	/*
	   Step 1: Check if at least one label missing
	*/

	if isAllTheLabelsSet(pod, all_candidates) {
		// The desired state and actual state of the Pod are the same.
		// No further action is required by the operator at this moment.
		log.FromContext(ctx).Info("no update required>>>" + fmt.Sprint(all_candidates))
		return ctrl.Result{}, nil
	}

	// If the label should be set but is not, set it.
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	for k, v := range all_candidates {
		pod.Labels[k] = v
	}

	log.FromContext(ctx).Info("adding labels ")

	/*
	   Step 2: Update the Pod in the Kubernetes API.
	*/

	if err := r.Update(ctx, &pod); err != nil {
		if apierrors.IsConflict(err) {
			// The Pod has been updated since we read it.
			// Requeue the Pod to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		if apierrors.IsNotFound(err) {
			// The Pod has been deleted since we read it.
			// Requeue the Pod to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		log.FromContext(ctx).Error(err, "unable to update Pod")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.GetLogger().Info("----- Env Dump -- ")
	mgr.GetLogger().Info("WATCHED_REGISTRY:" + watchedRegistry)
	mgr.GetLogger().Info("PREFIX_IMAGE_LABEL:" + prefixImageLabel)
	mgr.GetLogger().Info("PREFIX_POD_LABEL:" + prefixPodLabel)
	mgr.GetLogger().Info("REQUEST_DEBUG:" + getEnv("REQUEST_DEBUG", "false"))
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
