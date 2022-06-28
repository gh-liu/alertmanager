package js

import (
	"context"

	"github.com/dop251/goja"
	"github.com/go-kit/log"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

var scriptName = "alert.js"

type Runtime struct {
	conf   *config.JSConfig
	tmpl   *template.Template
	logger log.Logger

	rt      *goja.Runtime
	program *goja.Program
}

func New(conf *config.JSConfig, t *template.Template, l log.Logger) (*Runtime, error) {
	var r Runtime
	var err error
	program, err := goja.Compile(scriptName, conf.Script, true)
	if err != nil {
		return nil, err
	}
	r.program = program

	rt := goja.New()
	rt.Set("el", NewExtendLib())
	r.rt = rt

	r.tmpl = t
	r.logger = l
	r.conf = conf

	return &r, nil
}

func (n *Runtime) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	var err error

	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)
	body, err := n.tmpl.ExecuteTextString(n.conf.Text, data)
	if err != nil {
		return false, err
	}

	rt := n.rt
	rt.Set("message", body)
	rt.Set("targetID", n.conf.TargetID)
	_, err = rt.RunProgram(n.program)

	// 函数调用的方式
	// v := rt.Get("doHTTPRequest")
	// fn, ok := goja.AssertFunction(v)
	// if !ok {
	// 	return false, errors.New("doHTTPRequest function not found")
	// }
	//
	// alerts := rt.NewArray(as)
	// _, err := fn(goja.Undefined(), alerts)

	return false, err
}
