package js

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

type ExtendLib struct {
	extendMods map[string]interface{}
}

func NewExtendLib() *ExtendLib {
	extend := make(map[string]interface{})
	return &ExtendLib{extendMods: extend}
}

func (f *ExtendLib) HttpReq() *gorequest.SuperAgent {
	return gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

func (f *ExtendLib) Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (f *ExtendLib) AddExtendInfo(key string, value interface{}) {
	if f.extendMods == nil {
		f.extendMods = make(map[string]interface{})
	}
	f.extendMods[key] = value
}

func (f *ExtendLib) Microsecond() time.Duration {
	return time.Microsecond
}

func (f *ExtendLib) Millisecond() time.Duration {
	return time.Millisecond
}

func (f *ExtendLib) Second() time.Duration {
	return time.Second
}

func (f *ExtendLib) Minute() time.Duration {
	return time.Minute
}

func (f *ExtendLib) Hour() time.Duration {
	return time.Hour
}

func (f *ExtendLib) Sleep(t time.Duration) {
	time.Sleep(t)
}

func (f *ExtendLib) ToString(i int) string {
	return strconv.Itoa(i)
}

func (f *ExtendLib) SubString(s string, start, end int) string {
	if end < 0 {
		end = len(s)
	}
	if start > end {
		return ""
	}

	return s[start:end]
}

func (f *ExtendLib) Info(v interface{}) {
	log.Default().Println(v)
}

func (f *ExtendLib) Trace(v interface{}) {
	log.Default().Println(v)
}

func (f *ExtendLib) Debug(v interface{}) {
	log.Default().Println(v)
}

func (f *ExtendLib) Warn(v interface{}) {
	log.Default().Println(v)
}

func (f *ExtendLib) Error(v interface{}) {
	log.Default().Println(v)
}

func (f *ExtendLib) HmacSHA1(key, data string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (f *ExtendLib) CompressIp(ip string) string {
	address := net.ParseIP(ip)
	if address != nil {
		return address.String()
	}

	return ip
}
