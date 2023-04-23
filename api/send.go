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

package api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"net/http"

	"github.com/dop251/goja"
	"github.com/gh-liu/alertmanager/notify/js"
	"gopkg.in/gomail.v2"
)

var (
	SMTPFrom         string
	SMTPHello        string
	SMTPHost         string
	SMTPPort         int
	SMTPAuthUsername string
	SMTPAuthPassword string
	SMTPRequireTls   bool

	Script string
)

func (api *API) Add(mux *http.ServeMux) {
	mux.Handle("/send_mail", http.HandlerFunc(sendMail))
	mux.Handle("/send_syslog", http.HandlerFunc(sendSyslog))
	mux.Handle("/send_customhook", http.HandlerFunc(sendJS))
}

type email struct {
	host string
	port int

	from     string
	password string

	Subject string `json:"subject"`

	To      string `json:"to"`
	Message string `json:"message"`
}

func decodeJSON(r io.Reader, obj interface{}) error {
	return json.NewDecoder(r).Decode(obj)
}

func (e email) send() error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/html", e.Message)

	d := gomail.NewDialer(e.host, e.port, e.from, e.password)
	if !SMTPRequireTls {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func sendMail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var mail email
		body := r.Body
		defer body.Close()
		decodeJSON(body, &mail)
		mail.from = SMTPFrom
		mail.password = SMTPAuthPassword
		mail.host = SMTPHost
		mail.port = SMTPPort

		err := mail.send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

		} else {
			b, _ := json.Marshal(mail)
			w.Write(b)
		}
	}
}

type syslogger struct {
	Network string `json:"network"`
	RAddr   string `json:"raddr"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func (s syslogger) send() error {
	w, err := syslog.Dial(s.Network, s.RAddr, syslog.LOG_WARNING|syslog.LOG_DAEMON, s.Tag)
	if err != nil {
		return fmt.Errorf("failed to connect to syslog: %w", err)
	}
	defer w.Close()
	_, err = w.Write([]byte(s.Message))
	if err != nil {
		return fmt.Errorf("failed to write to syslog: %w", err)
	}
	return nil
}

func sendSyslog(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var sl syslogger
		body := r.Body
		defer body.Close()
		decodeJSON(body, &sl)
		err := sl.send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

		} else {
			b, _ := json.Marshal(sl)
			w.Write(b)
		}
	}
}

type scripter struct {
	TargetID string `json:"target_id"`
	Message  string `json:"message"`
	script   string `json:"-"`
}

var ErrNotFoundFunc = errors.New("告警脚本错误：找不到onSendMsgScriptlet函数")

func (s scripter) send() error {
	vm := goja.New()
	el := js.NewExtendLib()
	err1 := vm.Set("el", el)
	if err1 != nil {
		el.Error(err1)
		return err1
	}
	_, err2 := vm.RunString(s.script)
	if err2 != nil {
		el.Error(err2)
		return err2
	}

	type Msg struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}
	var msg Msg
	msg.Message = s.Message
	msg.ID = s.TargetID

	var fn func(Msg) string
	ret := vm.Get("onSendMsgScriptlet")
	if ret == nil {
		el.Error(ErrNotFoundFunc)
		return ErrNotFoundFunc
	}
	err := vm.ExportTo(ret, &fn)
	if err != nil {
		el.Error(ErrNotFoundFunc)
		return ErrNotFoundFunc
	}
	fn(msg)
	return nil
}

func sendJS(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var s scripter
		body := r.Body
		defer body.Close()
		decodeJSON(body, &s)
		s.script = Script

		err := s.send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

		} else {
			b, _ := json.Marshal(s)
			w.Write(b)
		}
	}
}
