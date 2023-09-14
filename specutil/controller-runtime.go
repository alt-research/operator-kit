// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"context"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ObjectList interface {
	client.ObjectList
	GetItems() []client.Object
}

type Builder[O client.Object, L ObjectList] struct {
	builder *builder.Builder
	mgr     ctrl.Manager

	typ      O
	listType L
	err      error
}

func NewControllerManagedBy[O client.Object, L ObjectList](mgr ctrl.Manager, typ O, listType L) *Builder[O, L] {
	b := builder.ControllerManagedBy(mgr)
	b.For(typ)
	return &Builder[O, L]{builder: b, mgr: mgr, typ: typ, listType: listType}
}

func (b *Builder[O, L]) Complete(r reconcile.Reconciler) error {
	if b.err != nil {
		return b.err
	}
	return b.builder.Complete(r)
}

func (b *Builder[O, L]) Named(name string) *Builder[O, L] {
	if b.err != nil {
		return b
	}
	b.builder.Named(name)
	return b
}

func (b *Builder[O, L]) Owns(obj client.Object) *Builder[O, L] {
	if b.err != nil {
		return b
	}
	b.builder.Owns(obj)
	return b
}

func (b *Builder[O, L]) Watches(obj client.Object, eventhandler handler.EventHandler, opts ...builder.WatchesOption) *Builder[O, L] {
	if b.err != nil {
		return b
	}
	b.builder.Watches(obj, eventhandler, opts...)
	return b
}

func (b *Builder[O, L]) WatchIndexed(watchedType client.Object, field string, fn func(obj O) []string) *Builder[O, L] {
	if b.err != nil {
		return b
	}
	b.err = b.mgr.GetFieldIndexer().IndexField(context.Background(), b.typ, field, func(o client.Object) []string { return fn(o.(O)) })
	b.builder.Watches(
		watchedType,
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			list := b.listType.DeepCopyObject().(L)
			listOps := &client.ListOptions{Namespace: obj.GetNamespace(), FieldSelector: fields.OneTermEqualSelector(field, obj.GetName())}
			if err := b.mgr.GetClient().List(ctx, list, listOps); err != nil {
				return []reconcile.Request{}
			}
			items := list.GetItems()
			requests := make([]reconcile.Request, len(items))
			for i, item := range items {
				requests[i] = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}}
			}
			return requests
		}),
	)
	return b
}

func (b *Builder[O, L]) WatchOneToOne(watchedType client.Object, fn func(obj O) string) *Builder[O, L] {
	if b.err != nil {
		return b
	}
	b.builder.Watches(
		watchedType,
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			name := fn(obj.(O))
			if name == "" {
				return []reconcile.Request{}
			}
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: obj.GetNamespace(), Name: name}}}
		}),
	)
	return b
}
