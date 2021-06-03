package output

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// init mock
var (
	changeSpecialChar2Dot = func(str string) string {
		reg, _ := regexp.Compile(`[^0-9a-zA-Z]+`)
		ret := reg.ReplaceAllString(str, ".")
		return ret
	}

	mockFilteredMsgs = []map[string]string{
		{"Env": "UAT", "Title": "测试应用", "Date": "[2021-06-01 10:43:50,253]", "Msg": "log message for <Test1>, have fun!!", "TimeTemplate": "[2006-01-02 15:04:05,000]"},
		{"Env": "UAT", "Title": "测试应用", "Date": "2020/06/01 13:50:30.020", "Msg": "log message for <Test2>, have fun too!!", "TimeTemplate": "2006/01/02 15:04:05.000"},
	}
	mockFormatStr     = "%v环境-%v日志：\n\t日志时间：%v\n\t日志信息：%v"
	mockTimestamp1, _ = time.Parse(changeSpecialChar2Dot(mockFilteredMsgs[0]["TimeTemplate"]), changeSpecialChar2Dot(mockFilteredMsgs[0]["Date"]))
	mockTimestamp2, _ = time.Parse(changeSpecialChar2Dot(mockFilteredMsgs[1]["TimeTemplate"]), changeSpecialChar2Dot(mockFilteredMsgs[1]["Date"]))
	mockEvent1        = beat.Event{
		Timestamp: mockTimestamp1,
		Fields:    common.MapStr{"Msg": fmt.Sprintf(mockFormatStr, mockFilteredMsgs[0]["Env"], mockFilteredMsgs[0]["Title"], mockFilteredMsgs[0]["Date"], mockFilteredMsgs[0]["Msg"])},
	}
	mockEvent2 = beat.Event{
		Timestamp: mockTimestamp2,
		Fields:    common.MapStr{"Msg": fmt.Sprintf(mockFormatStr, mockFilteredMsgs[1]["Env"], mockFilteredMsgs[1]["Title"], mockFilteredMsgs[1]["Date"], mockFilteredMsgs[1]["Msg"])},
	}
)

// mock for testing dingTalkApiOutput
var (
	mockPublishEvent1 = &publisher.Event{
		Content: mockEvent1,
	}
	mockPublishEvent2 = &publisher.Event{
		Content: mockEvent2,
	}

	mockConfig = config{
		URL:                   defaultConfig.URL,
		Codec:                 defaultConfig.Codec,
		MaxRetries:            defaultConfig.MaxRetries,
		Compression:           defaultConfig.Compression,
		KeepAlive:             defaultConfig.KeepAlive,
		MaxIdleConns:          defaultConfig.MaxIdleConns,
		IdleConnTimeout:       defaultConfig.IdleConnTimeout,
		ResponseHeaderTimeout: defaultConfig.ResponseHeaderTimeout,
		ApiAccessToken:        "", // TODO: 提交时取消
		At:                    defaultConfig.At,
		SendMsgType:           defaultConfig.SendMsgType,
		AddFields:             defaultConfig.AddFields,
	}

	mockHttpTransport = &http.Transport{
		MaxIdleConns:          mockConfig.MaxIdleConns,
		ResponseHeaderTimeout: time.Duration(mockConfig.ResponseHeaderTimeout) * time.Millisecond,
		IdleConnTimeout:       time.Duration(mockConfig.IdleConnTimeout) * time.Second,
		DisableCompression:    !mockConfig.Compression,
		DisableKeepAlives:     !mockConfig.KeepAlive,
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

	mockHttpClient = &http.Client{
		Transport: mockHttpTransport,
	}

	mockPool = sync.Pool{
		New: func() interface{} {
			req, err := http.NewRequest("POST", mockConfig.URL, nil)
			if err != nil {
				return err
			}
			return req
		},
	}
	mockMsg1, _ = mockEvent1.Fields.GetValue("Msg")
	mockText1   = TextMsg{
		Msgtype: TextMsgType,
		Text:    Text{Content: mockMsg1},
	}
	mockSerializedEvent1, _ = json.Marshal(&mockText1)
)

type mockObserver struct {
}

func (m mockObserver) NewBatch(i int) {
	return
}

func (m mockObserver) Acked(i int) {
	return
}

func (m mockObserver) Failed(i int) {
	return
}

func (m mockObserver) Dropped(i int) {
	return
}

func (m mockObserver) Duplicate(i int) {
	return
}

func (m mockObserver) Cancelled(i int) {
	return
}

func (m mockObserver) WriteError(err error) {
	return
}

func (m mockObserver) WriteBytes(i int) {
	return
}

func (m mockObserver) ReadError(err error) {
	return
}

func (m mockObserver) ReadBytes(i int) {
	return
}

func (m mockObserver) ErrTooMany(i int) {
	return
}

type mockBatch struct {
}

func (m mockBatch) Events() []publisher.Event {
	return []publisher.Event{*mockPublishEvent1, *mockPublishEvent2}
}

func (m mockBatch) ACK() {
	return
}

func (m mockBatch) Drop() {
	return
}

func (m mockBatch) Retry() {
	return
}

func (m mockBatch) RetryEvents(events []publisher.Event) {
	return
}

func (m mockBatch) Cancelled() {
	return
}

func (m mockBatch) CancelledEvents(events []publisher.Event) {
	return
}
