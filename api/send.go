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

		mail.send()
		b, _ := json.Marshal(mail)
		w.Write(b)
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
		sl.send()
		b, _ := json.Marshal(sl)
		w.Write(b)
	}
}

type scripter struct {
	TargetID string `json:"target_id"`
	Message  string `json:"message"`
	script   string `json:"-"`
}

func (s scripter) send() {
	program, err := goja.Compile("send.js", s.script, true)
	if err != nil {
		return
	}
	rt := goja.New()
	rt.Set("el", js.NewExtendLib())
	rt.Set("message", s.Message)
	rt.Set("targetID", s.TargetID)
	_, err = rt.RunProgram(program)
	if err != nil {
		return
	}
}

func sendJS(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var s scripter
		body := r.Body
		defer body.Close()
		decodeJSON(body, &s)
		s.script = Script

		s.send()
		b, _ := json.Marshal(s)
		w.Write(b)
	}
}
