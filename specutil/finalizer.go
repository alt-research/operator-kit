// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DeletionCallback func() error

func Finalize(ctx context.Context, c client.Client, obj client.Object, finalizer string, deletionCb DeletionCallback) (exit bool, err error) {
	// delete
	if !obj.GetDeletionTimestamp().IsZero() {
		if deletionCb != nil {
			if err = deletionCb(); err != nil {
				return
			}
		}
		// remove finalizer
		if controllerutil.ContainsFinalizer(obj, finalizer) {
			controllerutil.RemoveFinalizer(obj, finalizer)
			if err = c.Update(ctx, obj); err != nil {
				return
			}
			log.FromContext(ctx).V(1).Info("removed finalizer " + finalizer + " from object")
		}
		return true, nil
	}
	// create
	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		controllerutil.AddFinalizer(obj, finalizer)
		if err = c.Update(ctx, obj); err != nil {
			return
		}
		log.FromContext(ctx).V(1).Info("added finalizer " + finalizer + " to object")
	}
	return false, nil
}
