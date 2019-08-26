package extendedjob

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ejv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedjob/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/config"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/ctxlog"
	vss "code.cloudfoundry.org/cf-operator/pkg/kube/util/versionedsecretstore"
)

var _ reconcile.Reconciler = &ErrandReconciler{}

// NewErrandReconciler returns a new reconciler for errand jobs.
func NewErrandReconciler(
	ctx context.Context,
	config *config.Config,
	mgr manager.Manager,
	f setOwnerReferenceFunc,
	store vss.VersionedSecretStore,
) reconcile.Reconciler {
	jc := NewJobCreator(mgr.GetClient(), mgr.GetScheme(), f, store)

	return &ErrandReconciler{
		ctx:               ctx,
		client:            mgr.GetClient(),
		config:            config,
		scheme:            mgr.GetScheme(),
		setOwnerReference: f,
		jobCreator:        jc,
	}
}

// ErrandReconciler implements the Reconciler interface.
type ErrandReconciler struct {
	ctx               context.Context
	client            client.Client
	config            *config.Config
	scheme            *runtime.Scheme
	setOwnerReference setOwnerReferenceFunc
	jobCreator        JobCreator
}

// Reconcile starts jobs for extended jobs of the type errand with Run being set to 'now' manually.
func (r *ErrandReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	eJob := &ejv1.ExtendedJob{}

	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	ctxlog.Info(ctx, "Reconciling errand job ", request.NamespacedName)
	if err := r.client.Get(ctx, request.NamespacedName, eJob); err != nil {
		if apierrors.IsNotFound(err) {
			// Do not requeue, extended job is probably deleted.
			ctxlog.Infof(ctx, "Failed to find extended job '%s', not retrying: %s", request.NamespacedName, err)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		ctxlog.Errorf(ctx, "Failed to get the extended job '%s': %s", request.NamespacedName, err)
		return reconcile.Result{}, err
	}

	if eJob.Spec.Trigger.Strategy == ejv1.TriggerNow {
		// Set Strategy back to manual for errand jobs.
		eJob.Spec.Trigger.Strategy = ejv1.TriggerManual
		if err := r.client.Update(ctx, eJob); err != nil {
			return reconcile.Result{}, ctxlog.WithEvent(eJob, "UpdateError").Errorf(ctx, "Failed to revert to 'trigger.strategy=manual' on job '%s': %s", eJob.Name, err)
		}
	}

	if retry, err := r.jobCreator.Create(ctx, *eJob); err != nil {
		return reconcile.Result{}, ctxlog.WithEvent(eJob, "CreateJobError").Errorf(ctx, "Failed to create job '%s': %s", eJob.Name, err)
	} else if retry {
		ctxlog.Infof(ctx, "Retrying to create job '%s'", eJob.Name)
		result := reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 5,
		}
		return result, nil
	}

	ctxlog.WithEvent(eJob, "CreateJob").Infof(ctx, "Created errand job for '%s'", eJob.Name)

	if eJob.Spec.Trigger.Strategy == ejv1.TriggerOnce {
		// Traverse Strategy into the final 'done' state.
		eJob.Spec.Trigger.Strategy = ejv1.TriggerDone
		if err := r.client.Update(ctx, eJob); err != nil {
			return reconcile.Result{Requeue: false}, ctxlog.WithEvent(eJob, "UpdateError").Errorf(ctx, "Failed to traverse to 'trigger.strategy=done' on job '%s': %s", eJob.Name, err)
		}
	}

	return reconcile.Result{}, nil
}
