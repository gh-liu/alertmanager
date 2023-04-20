package js

import (
	"context"
	"errors"

	"github.com/dop251/goja"
	"github.com/gh-liu/alertmanager/config"
	"github.com/gh-liu/alertmanager/notify"
	"github.com/gh-liu/alertmanager/template"
	"github.com/gh-liu/alertmanager/types"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var scriptName = "alert.js"

type Runtime struct {
	conf   *config.JSConfig
	tmpl   *template.Template
	logger log.Logger

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
	if body == "" {
		return false, errors.New("[js] body is empty")
	}

	go func(bd, id string) {
		// level.Info(n.logger).Log("js_msg_body", bd)

		rt := goja.New()
		err3 := rt.Set("el", NewExtendLib())
		if err3 != nil {
			level.Error(n.logger).Log("error", err)
			return
		}

		_, err = rt.RunString(n.conf.Script)
		if err != nil {
			level.Error(n.logger).Log("error", err)
			return
		}

		var fn func(body, id string) string
		err2 := rt.ExportTo(rt.Get("sendMsg"), &fn)
		if err2 != nil {
			level.Error(n.logger).Log("error", err2)
			return
		}

		fn(bd, id)
		level.Info(n.logger).Log("js_msg_body_after", bd)

		// rt.Set("message", bd)
		// rt.Set("targetID", id)
		// _, err = rt.RunProgram(n.program)
	}(body, n.conf.TargetID)

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
