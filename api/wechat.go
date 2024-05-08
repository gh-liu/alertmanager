package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gh-liu/alertmanager/notify"
)

type Client struct {
	client *http.Client
	url    *url.URL

	APISecret string
	CorpID    string

	accessToken         string
	accessTokenExpireAt time.Time
}

func NewClient(apiurl, apiSecret string, corpID string) *Client {
	url, _ := url.Parse(apiurl)
	return &Client{
		client:    http.DefaultClient,
		url:       url,
		APISecret: apiSecret,
		CorpID:    corpID,
	}
}

type token struct {
	AccessToken string `json:"access_token"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	ExpiresInSec int64 `json:"expires_in"`
}

// GetWechatToken .
// https://developer.work.weixin.qq.com/document/path/91039
func (c *Client) GetWechatToken(ctx context.Context) (string, int64, error) {
	parameters := url.Values{}
	parameters.Add("corpsecret", c.APISecret)
	parameters.Add("corpid", c.CorpID)

	// copy url
	url2 := *c.url

	url2.Path += "gettoken"
	url2.RawQuery = parameters.Encode()

	resp, err := notify.Get(ctx, c.client, url2.String())
	if err != nil {
		return "", 0, notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	// b, _ := io.ReadAll(resp.Body)
	// r := string(b)
	// println(r)
	// return "", 0, nil

	if resp.StatusCode != 200 {
		return "", 0, fmt.Errorf("unexpected status code %v", resp.StatusCode)
	}

	var wechatToken token
	if err := json.NewDecoder(resp.Body).Decode(&wechatToken); err != nil {
		return "", 0, err
	}

	if wechatToken.ErrCode != 0 {
		return "", 0, errors.New(wechatToken.ErrMsg)
	}

	if wechatToken.AccessToken == "" {
		return "", 0, fmt.Errorf("invalid APISecret for CorpID: %s", c.CorpID)
	}

	return wechatToken.AccessToken, wechatToken.ExpiresInSec, nil
}

type weChatResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type weChatMessage struct {
	Text     weChatMessageContent `yaml:"text,omitempty" json:"text,omitempty"`
	ToUser   string               `yaml:"touser,omitempty" json:"touser,omitempty"`
	ToParty  string               `yaml:"toparty,omitempty" json:"toparty,omitempty"`
	Totag    string               `yaml:"totag,omitempty" json:"totag,omitempty"`
	AgentID  string               `yaml:"agentid,omitempty" json:"agentid,omitempty"`
	Safe     string               `yaml:"safe,omitempty" json:"safe,omitempty"`
	Type     string               `yaml:"msgtype,omitempty" json:"msgtype,omitempty"`
	Markdown weChatMessageContent `yaml:"markdown,omitempty" json:"markdown,omitempty"`
}

type weChatMessageContent struct {
	Content string `json:"content"`
}

// SendToWechat .
// https://developer.work.weixin.qq.com/document/path/90236
func (c *Client) SendToWechat(ctx context.Context, agentid, messageType, message, toUser, toParty, toTag string) error {
	// Refresh AccessToken after token expired
	if c.accessToken == "" || time.Now().After(c.accessTokenExpireAt) {
		token, expires, err := c.GetWechatToken(context.Background())
		if err != nil {
			return err
		}
		c.accessToken = token
		c.accessTokenExpireAt = time.Now().Add(time.Duration(expires) * time.Second)
	}

	msg := &weChatMessage{
		ToUser:  toUser,
		ToParty: toParty,
		Totag:   toTag,
		AgentID: agentid,
		Type:    messageType,
		Safe:    "0",
	}
	msgContent := weChatMessageContent{Content: message}
	if msg.Type == "markdown" {
		msg.Markdown = msgContent
	} else {
		msg.Text = msgContent
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	// copy url
	url2 := *c.url
	url2.Path += "message/send"
	q := url2.Query()
	q.Set("access_token", c.accessToken)
	url2.RawQuery = q.Encode()

	resp, err := notify.PostJSON(ctx, c.client, url2.String(), &buf)
	if err != nil {
		return notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var weResp weChatResponse
	if err := json.Unmarshal(body, &weResp); err != nil {
		return err
	}

	// https://work.weixin.qq.com/api/doc#10649
	switch weResp.ErrCode {
	case 0:
		return nil
	case 42001:
		// AccessToken is expired
		c.accessToken = ""
		return errors.New(weResp.ErrMsg)
	default:
		return errors.New(weResp.ErrMsg)
	}
}
