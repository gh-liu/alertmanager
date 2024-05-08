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
	var err2 error

	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)
	body, err2 := n.tmpl.ExecuteTextString(n.conf.Text, data)
	if err2 != nil {
		return false, err2
	}
	if body == "" {
		return false, errors.New("[js] body is empty")
	}

	go func(bd, id string) {
		defer func() {
			if r := recover(); r != nil {
				level.Error(n.logger).Log("recover error:", r)
			}
		}()

		vm := goja.New()
		err1 := vm.Set("el", NewExtendLib())
		if err1 != nil {
			level.Error(n.logger).Log("error", err1)
			return
		}
		_, err2 = vm.RunString(n.conf.Script)
		if err2 != nil {
			level.Error(n.logger).Log("error", err2)
			return
		}

		type Msg struct {
			Message string `json:"message"`
			ID      string `json:"id"`
		}
		var msg Msg
		msg.Message = bd
		msg.ID = id

		var fn func(Msg) string
		err := vm.ExportTo(vm.Get("onSendMsgScriptlet"), &fn)
		if err != nil {
			level.Error(n.logger).Log("error", "告警脚本错误：找不到onSendMsgScriptlet函数")
			return
		}
		fn(msg)
	}(body, n.conf.TargetID)

	return false, nil
}
