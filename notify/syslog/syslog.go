// Copyright 2015 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syslog

import (
	"context"
	"fmt"

	syslog "github.com/RackSec/srslog"
	"github.com/go-kit/log"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

type Syslog struct {
	conf   *config.SyslogConfig
	logger log.Logger
	tmpl   *template.Template
}

func New(c *config.SyslogConfig, t *template.Template, l log.Logger) (*Syslog, error) {
	sl := Syslog{
		logger: l,
		tmpl:   t,
		conf:   c,
	}
	return &sl, nil
}

func (sy *Syslog) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	data := notify.GetTemplateData(ctx, sy.tmpl, as, sy.logger)
	body, err := sy.tmpl.ExecuteTextString(sy.conf.Text, data)
	if err != nil {
		return false, err
	}

	w, err := syslog.Dial(sy.conf.Network, sy.conf.RAddr, syslog.LOG_WARNING|syslog.LOG_DAEMON, sy.conf.Tag)
	if err != nil {
		return false, fmt.Errorf("failed to connect to syslog: %w", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(body))
	if err != nil {
		return false, fmt.Errorf("failed to send to syslog: %w", err)
	}
	return true, nil
}
