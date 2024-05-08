package api_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gh-liu/alertmanager/api"
)

var client = api.NewClient(
	"https://qyapi.weixin.qq.com/cgi-bin/",
	os.Getenv("APISECRET"),
	os.Getenv("CORPID"),
)

func TestClientGetWechatToken(t *testing.T) {
	testCases := []struct {
		desc string
	}{
		{desc: "simple"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, i, err := client.GetWechatToken(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("s: %v\n", s)
			fmt.Printf("i: %v\n", i)
		})
	}
}

func TestClientSendToWechat(t *testing.T) {
	testCases := []struct {
		desc    string
		user    string
		agentid string
	}{
		{
			desc:    "liu",
			agentid: "1000026",
			user:    "lgh@gzsunrun.cn",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := client.SendToWechat(
				context.Background(),
				tC.agentid,
				"text",
				time.Now().String()+": test message",
				tC.user,
				"",
				"",
			)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
