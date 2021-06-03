package output

import (
	"context"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

func Test_dingTalkApiOutput_serializeTextMsgType(t *testing.T) {
	type fields struct {
		log       *logp.Logger
		beat      beat.Info
		observer  outputs.Observer
		codec     codec.Codec
		client    *http.Client
		serialize func(event *publisher.Event) ([]byte, error)
		reqPool   sync.Pool
		conf      config
	}
	type args struct {
		event *publisher.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "normal test",
			fields: fields{
				log: logp.NewLogger("dingTalkApi"),
				//beat:      mockBeatInfo,
				//observer:  mockObserver,
				//codec:     mockCodec,
				//client:    mockHttpClient,
				//serialize: nil,
				reqPool: mockPool,
				conf:    mockConfig,
			},
			args: args{event: mockPublishEvent1},
			want: mockSerializedEvent1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &dingTalkApiOutput{
				log:       tt.fields.log,
				beat:      tt.fields.beat,
				observer:  tt.fields.observer,
				codec:     tt.fields.codec,
				client:    tt.fields.client,
				serialize: tt.fields.serialize,
				reqPool:   tt.fields.reqPool,
				conf:      tt.fields.conf,
			}
			got, err := out.serializeTextMsgType(tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("serializeTextMsgType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serializeTextMsgType() got = %v, want %v", got, tt.want)
			} else {
				t.Log(string(got))
			}
		})
	}
}

func Test_dingTalkApiOutput_Publish(t *testing.T) {
	type fields struct {
		log       *logp.Logger
		beat      beat.Info
		observer  outputs.Observer
		codec     codec.Codec
		client    *http.Client
		serialize func(event *publisher.Event) ([]byte, error)
		reqPool   sync.Pool
		conf      config
	}
	type args struct {
		ctx   context.Context
		batch publisher.Batch
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			fields: fields{
				log:      logp.NewLogger("dingTalkApi"),
				observer: new(mockObserver),
				client:   mockHttpClient,
				reqPool:  mockPool,
				conf:     mockConfig,
			},
			args: args{batch: new(mockBatch)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &dingTalkApiOutput{
				log:       tt.fields.log,
				beat:      tt.fields.beat,
				observer:  tt.fields.observer,
				codec:     tt.fields.codec,
				client:    tt.fields.client,
				serialize: tt.fields.serialize,
				reqPool:   tt.fields.reqPool,
				conf:      tt.fields.conf,
			}
			out.serialize = out.serializeTextMsgType
			if err := out.Publish(tt.args.ctx, tt.args.batch); (err != nil) != tt.wantErr {
				t.Errorf("Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
