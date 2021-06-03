package output

import (
	"errors"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"net/http"
)

const (
	TextMsgType = "text"
)

type config struct {
	URL                   string                 `config:"url"`
	Codec                 codec.Config           `config:"codec"`
	MaxRetries            int                    `config:"max_retries"`
	Compression           bool                   `config:"compression"`
	KeepAlive             bool                   `config:"keep_alive"`
	MaxIdleConns          int                    `config:"max_idle_conns"`
	IdleConnTimeout       int                    `config:"idle_conn_timeout"`
	ResponseHeaderTimeout int                    `config:"response_header_timeout"`
	ApiAccessToken        string                 `config:"api_access_token"`
	At                    At                     `config:"at"`
	SendMsgType           string                 `config:"send_msg_type"`
	AddFields             map[string]interface{} `config:"add_fields"`
}

var (
	defaultConfig = config{
		URL:                   "https://oapi.dingtalk.com/robot/send",
		MaxRetries:            -1,
		Compression:           false,
		KeepAlive:             true,
		MaxIdleConns:          1,
		IdleConnTimeout:       0,
		ResponseHeaderTimeout: 3000,
		ApiAccessToken:        "",
		SendMsgType:           TextMsgType,
		AddFields:             make(map[string]interface{}, 0),
	}
)

func (c *config) Validate() error {
	_, err := http.NewRequest("POST", c.URL, nil)
	if err != nil {
		return err
	}
	if c.MaxIdleConns < 1 {
		return errors.New("max_idle_conns can't be <1")
	}
	if c.IdleConnTimeout < 0 {
		return errors.New("idle_conn_timeout can't be <0")
	}
	if c.ResponseHeaderTimeout < 1 {
		return errors.New("response_header_timeout can't be <1")
	}
	if c.ApiAccessToken == "" {
		return errors.New("api_token can't be empty")
	}
	return nil
}
