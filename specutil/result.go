package specutil

import (
	"fmt"
	"strings"

	"github.com/alt-research/operator-kit/array"
	. "github.com/alt-research/operator-kit/commonspec"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConditionResult represents an error with a status condition reason attached.
type ConditionResult struct {
	Type    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string

	Phase  PhaseType
	Exit   bool
	Result reconcile.Result
	Err    error

	noEvent bool
}

func (c ConditionResult) Error() string {
	return array.SliceStr(c.Err.Error(), ":20000")
}

// returns underlying error, used by errors package
func (c ConditionResult) Unwrap() error {
	return c.Err
}

func (c ConditionResult) AsCondition(typ ...string) metav1.Condition {
	t := c.Type
	if t == "" && len(typ) > 0 {
		t = typ[0]
	}
	msg := c.Message
	if msg == "" {
		msg = c.Err.Error()
	}
	return metav1.Condition{
		Type:    t,
		Status:  c.Status,
		Reason:  c.Reason,
		Message: msg,
	}
}

func (c *ConditionResult) WithExit(rst ...reconcile.Result) *ConditionResult {
	if c == nil {
		return nil
	}
	c.Exit = true
	if len(rst) > 0 {
		c.Result = rst[0]
	}
	return c
}

func (c *ConditionResult) WithType(typ string) *ConditionResult {
	if c == nil {
		return nil
	}
	c.Type = typ
	return c
}

func (c *ConditionResult) WithPhase(phase PhaseType) *ConditionResult {
	if c == nil {
		return nil
	}
	c.Phase = phase
	return c
}

func (c *ConditionResult) WithoutEvent() *ConditionResult {
	if c == nil {
		return nil
	}
	c.noEvent = true
	return c
}

func ConditionFail(reason string, msgOrErr any, a ...any) *ConditionResult {
	if msgOrErr == nil {
		return nil
	}
	if strings.Contains(reason, " ") {
		panic(fmt.Errorf("reason %q must not contain spaces", reason))
	}
	rst := &ConditionResult{Reason: reason, Status: metav1.ConditionFalse}
	switch msgOrErr := msgOrErr.(type) {
	case string:
		s := fmt.Sprintf(msgOrErr, a...)
		rst.Message = array.SliceStr(s, ":10000")
		rst.Err = errors.New(rst.Message)
	case error:
		if len(a) > 1 {
			s := fmt.Sprintf(a[0].(string), a[1:]...)
			rst.Err = errors.Wrap(msgOrErr, array.SliceStr(s, ":10000"))
		} else if len(a) > 0 {
			rst.Err = errors.Wrap(msgOrErr, array.SliceStr(a[0].(string), ":10000"))
		} else {
			rst.Err = msgOrErr
		}
		rst.Message = rst.Err.Error()
	default:
		rst.Message = fmt.Sprintf("%v", msgOrErr)
		rst.Err = errors.New(rst.Message)
	}
	return rst
}

func ConditionSuccess(reason string, msg string, args ...any) *ConditionResult {
	if strings.Contains(reason, " ") {
		panic(fmt.Errorf("reason %q must not contain spaces", reason))
	}
	rst := &ConditionResult{Reason: reason, Status: metav1.ConditionTrue, Message: fmt.Sprintf(msg, args...)}
	return rst
}

func ConditionUnknown(reason string, msg string, args ...any) *ConditionResult {
	if strings.Contains(reason, " ") {
		panic(fmt.Errorf("reason %q must not contain spaces", reason))
	}
	rst := &ConditionResult{Reason: reason, Status: metav1.ConditionUnknown, Message: fmt.Sprintf(msg, args...)}
	return rst
}
