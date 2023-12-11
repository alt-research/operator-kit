// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

import (
	"github.com/Masterminds/semver/v3"
	"github.com/containers/image/v5/docker/reference"
	corev1 "k8s.io/api/core/v1"
)

type ImageRef struct {
	//+optional
	Repository string `json:"repository"`
	//+optional
	Tag string `json:"tag,omitempty"`
	//+kubebuilder:validation:Pattern=`^sha256:[a-f0-9]{64}$`
	ID string `json:"id,omitempty"`
	//+kubebuilder:validation:Pattern=`^sha256:[a-f0-9]{64}$`
	Digest string `json:"digest,omitempty"`
	//+kubebuilder:default:=IfNotPresent
	//+kubebuilder:validation:Enum=Always;Never;IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
	//+optional
	//+patchMergeKey=name
	//+patchStrategy=merge
	PullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

func (r ImageRef) Ref() string {
	if r.Repository == "" {
		return ""
	}
	if r.Tag == "" {
		return r.Repository + ":latest"
	}
	return r.Repository + ":" + r.Tag
}

func (r ImageRef) RefDigest() string {
	if r.Digest == "" {
		return r.Ref()
	}
	return r.Repository + "@" + r.Digest
}

// returns true if the version of the image matches the given constraints
// if the version is invalid, returns nil
func (r ImageRef) VersionMatch(constraints string) *bool {
	con, err := semver.NewConstraint(constraints)
	if err != nil {
		return nil
	}
	ver, err := r.SemVer()
	if err != nil {
		return nil
	}
	t := con.Check(ver)
	return &t
}

func (r ImageRef) SemVer() (*semver.Version, error) {
	return semver.NewVersion(r.Tag)
}

func ImageRefFromRef(ref string) (ImageRef, error) {
	r, err := reference.Parse(ref)
	if err != nil {
		return ImageRef{}, err
	}
	ir := ImageRef{}
	if named, ok := r.(reference.Named); ok {
		ir.Repository = named.Name()
	}
	if tagged, ok := r.(reference.Tagged); ok {
		ir.Tag = tagged.Tag()
	}
	if digested, ok := r.(reference.Digested); ok {
		ir.Digest = digested.Digest().String()
	}
	return ir, nil
}
