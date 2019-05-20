/*
Copyright 2019 Rancher Labs, Inc.

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"

	"github.com/rancher/wrangler/pkg/generic"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	informers "k8s.io/client-go/informers/batch/v1"
	clientset "k8s.io/client-go/kubernetes/typed/batch/v1"
	listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
)

type JobHandler func(string, *v1.Job) (*v1.Job, error)

type JobController interface {
	JobClient

	OnChange(ctx context.Context, name string, sync JobHandler)
	OnRemove(ctx context.Context, name string, sync JobHandler)
	Enqueue(namespace, name string)

	Cache() JobCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type JobClient interface {
	Create(*v1.Job) (*v1.Job, error)
	Update(*v1.Job) (*v1.Job, error)
	UpdateStatus(*v1.Job) (*v1.Job, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Job, error)
	List(namespace string, opts metav1.ListOptions) (*v1.JobList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Job, err error)
}

type JobCache interface {
	Get(namespace, name string) (*v1.Job, error)
	List(namespace string, selector labels.Selector) ([]*v1.Job, error)

	AddIndexer(indexName string, indexer JobIndexer)
	GetByIndex(indexName, key string) ([]*v1.Job, error)
}

type JobIndexer func(obj *v1.Job) ([]string, error)

type jobController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.JobsGetter
	informer          informers.JobInformer
	gvk               schema.GroupVersionKind
}

func NewJobController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.JobsGetter, informer informers.JobInformer) JobController {
	return &jobController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromJobHandlerToHandler(sync JobHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Job
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Job))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *jobController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Job))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateJobOnChange(updater generic.Updater, handler JobHandler) JobHandler {
	return func(key string, obj *v1.Job) (*v1.Job, error) {
		if obj == nil {
			return handler(key, nil)
		}

		copyObj := obj.DeepCopy()
		newObj, err := handler(key, copyObj)
		if newObj != nil {
			copyObj = newObj
		}
		if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
			newObj, err := updater(copyObj)
			if newObj != nil && err == nil {
				copyObj = newObj.(*v1.Job)
			}
		}

		return copyObj, err
	}
}

func (c *jobController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *jobController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *jobController) OnChange(ctx context.Context, name string, sync JobHandler) {
	c.AddGenericHandler(ctx, name, FromJobHandlerToHandler(sync))
}

func (c *jobController) OnRemove(ctx context.Context, name string, sync JobHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromJobHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *jobController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *jobController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *jobController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *jobController) Cache() JobCache {
	return &jobCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *jobController) Create(obj *v1.Job) (*v1.Job, error) {
	return c.clientGetter.Jobs(obj.Namespace).Create(obj)
}

func (c *jobController) Update(obj *v1.Job) (*v1.Job, error) {
	return c.clientGetter.Jobs(obj.Namespace).Update(obj)
}

func (c *jobController) UpdateStatus(obj *v1.Job) (*v1.Job, error) {
	return c.clientGetter.Jobs(obj.Namespace).UpdateStatus(obj)
}

func (c *jobController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.Jobs(namespace).Delete(name, options)
}

func (c *jobController) Get(namespace, name string, options metav1.GetOptions) (*v1.Job, error) {
	return c.clientGetter.Jobs(namespace).Get(name, options)
}

func (c *jobController) List(namespace string, opts metav1.ListOptions) (*v1.JobList, error) {
	return c.clientGetter.Jobs(namespace).List(opts)
}

func (c *jobController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.Jobs(namespace).Watch(opts)
}

func (c *jobController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Job, err error) {
	return c.clientGetter.Jobs(namespace).Patch(name, pt, data, subresources...)
}

type jobCache struct {
	lister  listers.JobLister
	indexer cache.Indexer
}

func (c *jobCache) Get(namespace, name string) (*v1.Job, error) {
	return c.lister.Jobs(namespace).Get(name)
}

func (c *jobCache) List(namespace string, selector labels.Selector) ([]*v1.Job, error) {
	return c.lister.Jobs(namespace).List(selector)
}

func (c *jobCache) AddIndexer(indexName string, indexer JobIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Job))
		},
	}))
}

func (c *jobCache) GetByIndex(indexName, key string) (result []*v1.Job, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.Job))
	}
	return result, nil
}
