/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"
	"net/smtp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.Log.WithValues("pod", req.NamespacedName)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since we can get them on deleted requests.
			emailErr := r.sendEmailNotification(&pod, "deleted")
			if emailErr != nil {
				return ctrl.Result{}, emailErr
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	emailErr := r.sendEmailNotification(&pod, "created")
	if emailErr != nil {
		return ctrl.Result{}, emailErr
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func (r *PodReconciler) sendEmailNotification(pod *corev1.Pod, action string) error {
	// Send email notification
	subject := fmt.Sprintf("Pod %s: %s", action, pod.Name)
	body := fmt.Sprintf("Pod %s %s in namespace %s", action, pod.Name, pod.Namespace)
	err := sendEmail(subject, body)
	if err != nil {
		return err
	}

	return nil
}

func sendEmail(subject, body string) error {
	from := "namnt194337@gmail.com"
	password := "Nam12345"
	to := "trieu81200111@gmail.com"

	// Set up authentication information.
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	// Compose the message.
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Send the message.
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
