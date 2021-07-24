/*
Copyright 2021 The Crossplane Authors.

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

package datasource

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/benagricola/provider-externaldata/apis/datasource/v1alpha1"
	apisv1alpha1 "github.com/benagricola/provider-externaldata/apis/v1alpha1"
)

const (
	errNotDataSource = "managed resource is not a DataSource custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetPC         = "cannot get ProviderConfig"

	errConfigMapName = "configMapName must be specified when type is configmap"
	errURI           = "uri must be specified when type is uri"
	errDataLookup    = "cannot retrieve from datasource"

	errFmtUnknownSourceType = "unknown datasource type %s"
	errFmtRequestFailed     = "request failed: %s"
)

// Setup adds a controller that reconciles DataSource managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.DataSourceGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DataSourceGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.DataSource{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return nil, errors.New(errNotDataSource)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	return &external{
		client: c.kube,
		ns:     pc.Spec.Namespace,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client client.Client
	ns     string
}

func lookupConfigMap(ctx context.Context, client client.Client, namespace string, name string, re *runtime.RawExtension) error { //nolint:interfacer
	// Interfacer linting disabled as it tries to suggest json.Unmarshaler
	cm := &apiv1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cm); err != nil {
		return err
	}
	mb, err := json.Marshal(cm.Data)
	if err != nil {
		return err
	}
	return re.UnmarshalJSON(mb)
}

func lookupURL(ctx context.Context, uri string, re *runtime.RawExtension) error { //nolint:interfacer
	// Interfacer linting disabled as it tries to suggest json.Unmarshaler
	c := resty.New()
	c.SetRetryCount(1)
	c.SetTimeout(1 * time.Second)
	c.SetHeader("Accept", "application/json")

	res, err := c.R().
		SetContext(ctx).
		Get(uri)

	if err != nil {
		return err
	}

	if !res.IsSuccess() {
		return errors.Errorf(errFmtRequestFailed, res.Status())
	}

	return re.UnmarshalJSON(res.Body())
}

func lookupData(ctx context.Context, client client.Client, ext external, sp v1alpha1.DataSourceSpec, re *runtime.RawExtension) error {
	var err error

	switch sp.ForProvider.SourceType {
	case v1alpha1.SourceTypeConfigMap:

		if sp.ForProvider.ConfigMapName == nil {
			return errors.New(errConfigMapName)
		}
		err = lookupConfigMap(ctx, client, ext.ns, *sp.ForProvider.ConfigMapName, re)

	case v1alpha1.SourceTypeURL:
		if sp.ForProvider.URL == nil {
			return errors.New(errURI)
		}
		err = lookupURL(ctx, *sp.ForProvider.URL, re)
	default:
		return errors.Errorf(errFmtUnknownSourceType, sp.ForProvider.SourceType)
	}

	return err
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDataSource)
	}

	// If deletion was requested, return that this resource does not exist
	// or the Kubernetes API object will not be deleted.
	if cr.DeletionTimestamp != nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	nd := runtime.RawExtension{}

	err := lookupData(
		ctx,
		c.client,
		*c,
		cr.Spec,
		&nd)

	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errDataLookup)
	}

	upToDate := cmp.Equal(cr.Status.AtProvider, &nd)

	return managed.ExternalObservation{
		ResourceExists:   cr.Status.AtProvider != nil,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDataSource)
	}

	nd := runtime.RawExtension{}

	err := lookupData(
		ctx,
		c.client,
		*c,
		cr.Spec,
		&nd)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errDataLookup)
	}

	cr.Status.AtProvider = &nd

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDataSource)
	}

	err := lookupData(
		ctx,
		c.client,
		*c,
		cr.Spec,
		cr.Status.AtProvider)

	return managed.ExternalUpdate{}, errors.Wrap(err, errDataLookup)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return errors.New(errNotDataSource)
	}
	cr.Status.AtProvider = nil

	return nil
}
