package output

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ClessLi/beats-output-ding-talk-api/resolver"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var dnsCache = resolver.NewDNSResolver()

func init() {
	outputs.RegisterType("ding_talk_api", makeDingTalkApi)
}

type dingTalkApiOutput struct {
	log       *logp.Logger
	beat      beat.Info
	observer  outputs.Observer
	codec     codec.Codec
	client    *http.Client
	serialize func(event *publisher.Event) ([]byte, error)
	reqPool   sync.Pool
	conf      config
}

// makeDingTalkApi instantiates a new dingTalk robot api output instance.
func makeDingTalkApi(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {

	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	ho := &dingTalkApiOutput{
		log:      logp.NewLogger("dingTalkApi"),
		beat:     beat,
		observer: observer,
		conf:     config,
	}

	// disable bulk support in publisher pipeline
	if err := cfg.SetInt("bulk_max_size", -1, -1); err != nil {
		ho.log.Error("Disable bulk error: ", err)
	}

	//select serializer
	switch ho.conf.SendMsgType {
	case TextMsgType:
		ho.serialize = ho.serializeTextMsgType
	}
	//ho.serialize = ho.serializeAll

	//if config.OnlyFields {
	//	ho.serialize = ho.serializeOnlyFields
	//}

	// init output
	if err := ho.init(beat, config); err != nil {
		return outputs.Fail(err)
	}

	return outputs.Success(-1, config.MaxRetries, ho)
}

func (out *dingTalkApiOutput) init(beat beat.Info, c config) error {
	var err error

	out.codec, err = codec.CreateEncoder(beat, c.Codec)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		MaxIdleConns:          out.conf.MaxIdleConns,
		ResponseHeaderTimeout: time.Duration(out.conf.ResponseHeaderTimeout) * time.Millisecond,
		IdleConnTimeout:       time.Duration(out.conf.IdleConnTimeout) * time.Second,
		DisableCompression:    !out.conf.Compression,
		DisableKeepAlives:     !out.conf.KeepAlive,
		DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := dnsCache.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				var dialer net.Dialer
				conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err == nil {
					break
				}
			}
			return
		},
	}

	out.client = &http.Client{
		Transport: tr,
	}

	out.reqPool = sync.Pool{
		New: func() interface{} {
			req, err := http.NewRequest("POST", out.conf.URL, nil)
			if err != nil {
				return err
			}
			return req
		},
	}

	out.log.Infof("Initialized dingTalk robot api output:\n"+
		"url=%v\n"+
		"codec=%v\n"+
		"max_retries=%v\n"+
		"compression=%v\n"+
		"keep_alive=%v\n"+
		"max_idle_conns=%v\n"+
		"idle_conn_timeout=%vs\n"+
		"response_header_timeout=%vms\n"+
		"api_access_token=%v\n"+
		"at=%v\n"+
		"send_msg_type=%v\n",
		c.URL, c.Codec, c.MaxRetries, c.Compression,
		c.KeepAlive, c.MaxIdleConns, c.IdleConnTimeout, c.ResponseHeaderTimeout,
		c.ApiAccessToken, c.At, c.SendMsgType)
	return nil
}

// Implement Client
func (out *dingTalkApiOutput) Close() error {
	out.client.CloseIdleConnections()
	return nil
}

//func (out *dingTalkApiOutput) serializeOnlyFields(event *publisher.Event) ([]byte, error) {
//	fields := event.Content.Fields
//	fields["@timestamp"] = event.Content.Timestamp
//	for key, val := range out.conf.AddFields {
//		fields[key] = val
//	}
//	serializedEvent, err := json.Marshal(&fields)
//	if err != nil {
//		out.log.Error("Serialization error: ", err)
//		return make([]byte, 0), err
//	}
//	return serializedEvent, nil
//}
//
//func (out *dingTalkApiOutput) serializeAll(event *publisher.Event) ([]byte, error) {
//	serializedEvent, err := out.codec.Encode(out.beat.Beat, &event.Content)
//	if err != nil {
//		out.log.Error("Serialization error: ", err)
//		return make([]byte, 0), err
//	}
//	return serializedEvent, nil
//}

func (out *dingTalkApiOutput) serializeTextMsgType(event *publisher.Event) ([]byte, error) {
	var err error
	textMsg := defaultTextMsg
	if out.conf.At.AtUserIds != nil || out.conf.At.AtMobiles != nil {
		textMsg.At = out.conf.At
	}
	textMsg.Text.Content, err = event.Content.Fields.GetValue("Msg")
	if err != nil {
		out.log.Warnf("Read event's message error: %v", err)
		return make([]byte, 0), err
	}
	serializedEvent, err := json.Marshal(&textMsg)
	if err != nil {
		out.log.Error("Serialization error: ", err)
		return make([]byte, 0), err
	}
	return serializedEvent, nil
}

func (out *dingTalkApiOutput) Publish(_ context.Context, batch publisher.Batch) error {
	st := out.observer
	events := batch.Events()
	st.NewBatch(len(events))

	if len(events) == 0 {
		batch.ACK()
		return nil
	}

	for i := range events {
		event := events[i]

		serializedEvent, err := out.serialize(&event)

		if err != nil {
			if event.Guaranteed() {
				out.log.Errorf("Failed to serialize the event: %+v", err)
			} else {
				out.log.Warnf("Failed to serialize the event: %+v", err)
			}
			out.log.Debugf("Failed event: %v", event)

			batch.RetryEvents(events)
			st.Failed(len(events))
			return nil
		}

		if err = out.send(serializedEvent); err != nil {
			if event.Guaranteed() {
				out.log.Errorf("Writing event to dingTalk robot api failed with: %+v", err)
			} else {
				out.log.Warnf("Writing event to dingTalk robot api failed with: %+v", err)
			}

			batch.RetryEvents(events)
			st.Failed(len(events))
			return nil
		}
	}

	batch.ACK()
	st.Acked(len(events))
	return nil
}

func (out *dingTalkApiOutput) String() string {
	return "dingTalkApi(" + out.conf.URL + ")"
}

func (out *dingTalkApiOutput) send(data []byte) error {

	req, err := out.getReq(data)
	if err != nil {
		return err
	}
	defer out.putReq(req)

	resp, err := out.client.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		out.log.Warnf("Read response result error: %v", err)
	}

	if len(respBody) > 0 {
		respBodyStr := string(respBody)
		out.log.Info(respBodyStr)
	}

	err = resp.Body.Close()
	if err != nil {
		out.log.Warn("Close response body error:", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	return nil
}

func (out *dingTalkApiOutput) getReq(data []byte) (*http.Request, error) {
	tmp := out.reqPool.Get()

	req, ok := tmp.(*http.Request)
	if ok {
		buf := bytes.NewBuffer(data)
		req.Body = ioutil.NopCloser(buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "beat "+out.beat.Version)
		//if out.conf.Username != "" {
		//	req.SetBasicAuth(out.conf.Username, out.conf.Password)
		//}
		params := req.URL.Query()
		params.Set("access_token", out.conf.ApiAccessToken)
		req.URL.RawQuery = params.Encode()
		return req, nil
	}

	err, ok := tmp.(error)
	if ok {
		return nil, err
	}

	return nil, errors.New("pool assertion error")
}

func (out *dingTalkApiOutput) putReq(req *http.Request) {
	out.reqPool.Put(req)
}
