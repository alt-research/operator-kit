// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/alt-research/operator-kit/commonspec"
	"github.com/alt-research/operator-kit/must"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var doNotCatchPanic = os.Getenv("DO_NOT_CATCH_PANIC") == "true"

type StepFunc func() error

type Step struct {
	ConditionType  string
	TransitionFunc StepFunc
	opts           PatchConditionOptions
}

type StepSkipper func(s Step, cond *metav1.Condition) bool

// ConditionManager is a helper to separate the logic of managing conditions from the controller logic.
type ConditionManager struct {
	client             client.Client
	req                reconcile.Request
	obj                client.Object
	cp                 *commonspec.ConditionPhase
	steps              []Step
	skipper            func() bool
	preFinalizeSkipper func() bool
	stepSkipper        StepSkipper
	finalizer          string
	finalizeFunc       func() error
	afterDeletion      func()
	defaultFailResult  reconcile.Result
	eventRecorder      record.EventRecorder
}

func NewConditionManager(client client.Client, req reconcile.Request, obj client.Object, cp *commonspec.ConditionPhase) *ConditionManager {
	return &ConditionManager{client: client, obj: obj, req: req, cp: cp, defaultFailResult: reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}}
}

func (m *ConditionManager) WithDefaultFailResult(result reconcile.Result) *ConditionManager {
	m.defaultFailResult = result
	return m
}

func (m *ConditionManager) WithEventRecorder(eventRecorder record.EventRecorder) *ConditionManager {
	m.eventRecorder = eventRecorder
	return m
}

// Runs after finalize, If skipper function returns true, all reconcilation steps will be skipped
func (m *ConditionManager) WithSkipper(f func() bool) *ConditionManager {
	if m.skipper != nil {
		panic("skipper already set")
	}
	m.skipper = f
	return m
}

// Runs before finalize, If skipper function returns true, all reconcilation steps including finalizing will be skipped
func (m *ConditionManager) WithPreFinalizeSkipper(f func() bool) *ConditionManager {
	if m.preFinalizeSkipper != nil {
		panic("preFinalizeSkipper already set")
	}
	m.preFinalizeSkipper = f
	return m
}

// Runs before every step, If skipper function returns true, step will be skipped
func (m *ConditionManager) WithStepSkipper(f StepSkipper) *ConditionManager {
	if m.stepSkipper != nil {
		panic("skipper already set")
	}
	m.stepSkipper = f
	return m
}

func (m *ConditionManager) WithFinalizer(finalizer string, f func() error) *ConditionManager {
	if m.finalizer != "" {
		panic("finalizer already set")
	}
	m.finalizer = finalizer
	m.finalizeFunc = f
	return m
}

// will be called when the reconcilation triggerd after resource deleted (Get returns NotFound)
func (m *ConditionManager) WithAfterDeletion(f func()) *ConditionManager {
	if m.afterDeletion != nil {
		panic("after deletion already set")
	}
	m.afterDeletion = f
	return m
}

func (m *ConditionManager) Step(conditionType string, f StepFunc, options ...PatchConditionOption) *ConditionManager {
	if conditionType == "" {
		panic("condition type cannot be empty")
	}
	if f == nil {
		panic("transition function cannot be nil")
	}
	opts := PatchConditionOptions{
		SuccessReason:            conditionType + "Succeeded",
		SuccessMessage:           conditionType + " Succeeded",
		DefaultFailReason:        conditionType + "Failed",
		DefaultFailMessagePrefix: conditionType + " Failed",
		AfterConditionSet:        func() error { return nil },
	}
	for _, opter := range options {
		opter(&opts)
	}
	m.steps = append(m.steps, Step{ConditionType: conditionType, TransitionFunc: f, opts: opts})
	return m
}

func (m *ConditionManager) Run(ctx context.Context) (reconcile.Result, error) {
	log := log.FromContext(ctx)
	if err := m.client.Get(ctx, m.req.NamespacedName, m.obj); err != nil {
		log.V(3).Error(err, "failed to get object")
		if m.afterDeletion != nil {
			m.afterDeletion()
		}
		return reconcile.Result{}, nil
	}
	if m.preFinalizeSkipper != nil && m.preFinalizeSkipper() {
		return reconcile.Result{}, nil
	}
	if m.finalizer != "" {
		if exit, err := Finalize(ctx, m.client, m.obj, m.finalizer, m.finalizeFunc); err != nil {
			log.Error(err, "failed to finalize")
			apimeta.SetStatusCondition(&m.cp.Conditions, metav1.Condition{
				Type:    "Finalizing",
				Status:  metav1.ConditionFalse,
				Reason:  "FinalizationFailed",
				Message: err.Error(),
			})
			m.cp.Phase = commonspec.PhaseFinalizationError
			if err := m.client.Status().Update(ctx, m.obj); err != nil {
				log.Error(err, "failed to update status")
			}
			return m.defaultFailResult, nil
		} else if exit {
			return reconcile.Result{}, nil
		}
	}
	if m.skipper != nil && m.skipper() {
		return reconcile.Result{}, nil
	}
	// Patch Object with condition
	for _, s := range m.steps {
		cond := apimeta.FindStatusCondition(m.cp.Conditions, s.ConditionType)
		if m.stepSkipper != nil && m.stepSkipper(s, cond) {
			continue
		}
		if s.opts.SetProcessing {
			apimeta.SetStatusCondition(&m.cp.Conditions, metav1.Condition{
				Type:    s.ConditionType,
				Status:  metav1.ConditionUnknown,
				Reason:  s.ConditionType + "Processing",
				Message: "Processing",
			})
			if err := s.opts.AfterConditionSet(); err != nil {
				log.Error(err, "failed to run after condition set")
				return must.Default(s.opts.FailResult, m.defaultFailResult), nil
			}
			if err := m.client.Status().Update(ctx, m.obj); err != nil {
				log.Error(err, "failed to update status")
				return must.Default(s.opts.FailResult, m.defaultFailResult), nil
			}
		}
		var exit bool
		var rst reconcile.Result
		if _, err := patch(ctx, m.client, m.obj, func() error {
			var panicErr error
			err := func() error {
				if !doNotCatchPanic {
					defer func() {
						r := recover()
						if r != nil {
							panicErr = fmt.Errorf("panic: %v\n%s", r, string(debug.Stack()))
						}
					}()
				}
				return s.TransitionFunc()
			}()
			if panicErr != nil {
				err = panicErr
			}
			// return nil and Success
			if err == nil || reflect.ValueOf(err).IsZero() {
				apimeta.SetStatusCondition(&m.cp.Conditions, metav1.Condition{
					Type:    s.ConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  s.opts.SuccessReason,
					Message: s.opts.SuccessMessage,
				})
				if err := s.opts.AfterConditionSet(); err != nil {
					return err
				}
				return nil
			}
			reason := s.opts.DefaultFailReason
			prefix := s.opts.DefaultFailMessagePrefix
			var phaseSet bool
			var cond metav1.Condition
			if e, ok := err.(*ConditionResult); ok {
				cond = e.AsCondition(s.ConditionType)
				exit = e.Exit
				rst = e.Result
				if e.Phase != "" {
					m.cp.Phase = e.Phase
					phaseSet = true
				}
				if !e.noEvent && m.eventRecorder != nil {
					typ := corev1.EventTypeNormal
					if cond.Status == metav1.ConditionFalse {
						typ = corev1.EventTypeWarning
					}
					m.eventRecorder.Event(m.obj, typ, cond.Reason, cond.Message)
				}
			} else {
				cond = metav1.Condition{
					Type:    s.ConditionType,
					Status:  metav1.ConditionFalse,
					Reason:  reason,
					Message: fmt.Sprintf("%s: %s", prefix, err.Error()),
				}
			}
			apimeta.SetStatusCondition(&m.cp.Conditions, cond)
			// condition failed
			if cond.Status == metav1.ConditionFalse {
				log.Info("condition failed", "reason", cond.Reason, "message", cond.Message)
				if !phaseSet {
					m.cp.Phase = commonspec.PhaseType(cond.Reason)
				}
				if s.opts.FailCb != nil {
					s.opts.FailCb()
				}
			} else {
				err = nil
			}
			if err := s.opts.AfterConditionSet(); err != nil {
				return err
			}
			return err
		}, false); err != nil {
			log.Error(err, "step failed", "step", s.ConditionType)
			return must.Default(rst, s.opts.FailResult, m.defaultFailResult), nil
		} else if exit {
			return must.Default(rst, s.opts.FailResult, m.defaultFailResult), nil
		}
	}
	return reconcile.Result{}, nil
}
