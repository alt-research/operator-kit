// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cu "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// mutate wraps a MutateFn and applies validation to its result.
func mutate(f cu.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}
	return nil
}

// Patch the given object in the Kubernetes
// cluster. The object's desired state must be reconciled with the before
// state inside the passed in callback MutateFn.
//
// It returns the executed operation and an error.
func Patch(ctx context.Context, c client.Client, obj client.Object, f cu.MutateFn) (cu.OperationResult, error) {
	return patch(ctx, c, obj, f, true, nil)
}

func patch(ctx context.Context, c client.Client, obj client.Object, f cu.MutateFn, abortOnMutateError bool, onChange func()) (cu.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		return cu.OperationResultNone, err
	}

	// Create patches for the object and its possible status.
	objPatch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	statusPatch := client.MergeFrom(obj.DeepCopyObject().(client.Object))

	// Create a copy of the original object as well as converting that copy to
	// unstructured data.
	before, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj.DeepCopyObject())
	if err != nil {
		return cu.OperationResultNone, err
	}

	// Attempt to extract the status from the resource for easier comparison later
	beforeStatus, hasBeforeStatus, err := unstructured.NestedFieldCopy(before, "status")
	if err != nil {
		return cu.OperationResultNone, err
	}

	// If the resource contains a status then remove it from the unstructured
	// copy to avoid unnecessary patching later.
	if hasBeforeStatus {
		unstructured.RemoveNestedField(before, "status")
	}

	// Mutate the original object.
	var mutateErr error
	if mutateErr = mutate(f, key, obj); mutateErr != nil && abortOnMutateError {
		return cu.OperationResultNone, err
	}

	// Convert the resource to unstructured to compare against our before copy.
	after, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return cu.OperationResultNone, err
	}

	// Attempt to extract the status from the resource for easier comparison later
	afterStatus, hasAfterStatus, err := unstructured.NestedFieldCopy(after, "status")
	if err != nil {
		return cu.OperationResultNone, err
	}

	// If the resource contains a status then remove it from the unstructured
	// copy to avoid unnecessary patching later.
	if hasAfterStatus {
		unstructured.RemoveNestedField(after, "status")
	}

	if !reflect.DeepEqual(before, after) || ((hasBeforeStatus || hasAfterStatus) && !reflect.DeepEqual(beforeStatus, afterStatus)) {
		if onChange != nil {
			// make additional change
			onChange()
			// calculate new afters
			// Convert the resource to unstructured to compare against our before copy.
			after, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				return cu.OperationResultNone, err
			}

			// Attempt to extract the status from the resource for easier comparison later
			afterStatus, hasAfterStatus, err = unstructured.NestedFieldCopy(after, "status")
			if err != nil {
				return cu.OperationResultNone, err
			}

			// If the resource contains a status then remove it from the unstructured
			// copy to avoid unnecessary patching later.
			if hasAfterStatus {
				unstructured.RemoveNestedField(after, "status")
			}
		}
	}

	result := cu.OperationResultNone

	if !reflect.DeepEqual(before, after) {
		// Only issue a Patch if the before and after resources (minus status) differ
		if err := c.Patch(ctx, obj, objPatch); err != nil {
			return result, err
		}
		result = cu.OperationResultUpdated
	}

	if (hasBeforeStatus || hasAfterStatus) && !reflect.DeepEqual(beforeStatus, afterStatus) {
		// Only issue a Status Patch if the resource has a status and the beforeStatus
		// and afterStatus copies differ
		if result == cu.OperationResultUpdated {
			// If Status was replaced by Patch before, set it to afterStatus
			objectAfterPatch, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				return result, err
			}
			if err = unstructured.SetNestedField(objectAfterPatch, afterStatus, "status"); err != nil {
				return result, err
			}
			// If Status was replaced by Patch before, restore patched structure to the obj
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(objectAfterPatch, obj); err != nil {
				return result, err
			}
		}
		if err := c.Status().Patch(ctx, obj, statusPatch); err != nil {
			return result, err
		}
		if result == cu.OperationResultUpdated {
			result = cu.OperationResultUpdatedStatus
		} else {
			result = cu.OperationResultUpdatedStatusOnly
		}
	}

	return result, mutateErr
}
