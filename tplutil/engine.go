// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package tplutil

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig/v3"
)

var FuncMap = template.FuncMap{}

func New(name string) *template.Template {
	return template.New(name).Funcs(sprig.FuncMap()).Funcs(FuncMap)
}

func Render(t *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	return buf.String(), err
}
