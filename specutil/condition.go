// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"context"
	"fmt"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PatchConditionOptions struct {
	SetProcessing            bool
	SuccessReason            string
	SuccessMessage           string
	DefaultFailReason        string
	DefaultFailMessagePrefix string
	AfterConditionSet        func() error
	FailCb                   func()
	FailResult               reconcile.Result
}

func PatchWithCondition(ctx context.Context, c client.Client, obj client.Object, conditions *[]metav1.Condition, conditionType string, procedure func() error, opts ...PatchConditionOption) (controllerutil.OperationResult, error) {
	opt := PatchConditionOptions{
		SuccessReason:            conditionType + "Succeeded",
		SuccessMessage:           conditionType + " Succeeded",
		DefaultFailReason:        conditionType + "Failed",
		DefaultFailMessagePrefix: conditionType + " Failed",
		AfterConditionSet:        func() error { return nil },
	}
	for _, opter := range opts {
		opter(&opt)
	}
	if opt.SetProcessing {
		apimeta.SetStatusCondition(conditions, metav1.Condition{
			Type:    conditionType,
			Status:  metav1.ConditionUnknown,
			Reason:  conditionType + "Processing",
			Message: "Processing",
		})
		if err := opt.AfterConditionSet(); err != nil {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Status().Update(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
	}
	return patch(ctx, c, obj, func() error {
		err := procedure()
		if err == nil {
			apimeta.SetStatusCondition(conditions, metav1.Condition{
				Type:    conditionType,
				Status:  metav1.ConditionTrue,
				Reason:  opt.SuccessReason,
				Message: opt.SuccessMessage,
			})
			if err := opt.AfterConditionSet(); err != nil {
				return err
			}
			return nil
		}
		reason := opt.DefaultFailReason
		prefix := opt.DefaultFailMessagePrefix
		if e, ok := err.(ConditionResult); ok {
			reason = e.Reason
		} else if e, ok := err.(*ConditionResult); ok {
			reason = e.Reason
		}
		apimeta.SetStatusCondition(conditions, metav1.Condition{
			Type:    conditionType,
			Status:  metav1.ConditionFalse,
			Reason:  reason,
			Message: fmt.Sprintf("%s: %s", prefix, err.Error()),
		})
		if err := opt.AfterConditionSet(); err != nil {
			return err
		}
		return err
	}, false, nil)
}

type PatchConditionOption func(opts *PatchConditionOptions)

func WithSuccessReason(successReason string) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.SuccessReason = successReason
	}
}

func WithSuccessMessage(SuccessMessage string) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.SuccessMessage = SuccessMessage
	}
}

func WithDefaultFailReason(DefaultFailReason string) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.DefaultFailReason = DefaultFailReason
	}
}

func WithDefaultFailMessagePrefix(DefaultFailMessagePrefix string) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.DefaultFailMessagePrefix = DefaultFailMessagePrefix
	}
}

func WithSetProcessing() PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.SetProcessing = true
	}
}

func WithoutSetProcessing() PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.SetProcessing = false
	}
}

func WithAfterConditionSet(f func() error) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.AfterConditionSet = f
	}
}

func WithFailResult(failResult reconcile.Result) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.FailResult = failResult
	}
}

func WithFailCb(failCb func()) PatchConditionOption {
	return func(opts *PatchConditionOptions) {
		opts.FailCb = failCb
	}
}

func LastTransitionTime(conditions *[]metav1.Condition) (t time.Time) {
	for _, c := range *conditions {
		if c.LastTransitionTime.After(t) {
			t = c.LastTransitionTime.Time
		}
	}
	return
}
