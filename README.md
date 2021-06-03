[![License](http://img.shields.io/badge/license-mit-blue.svg)](https://github.com/ClessLi/beats-output-ding-talk-api/master/LICENSE)
[![Build Status](https://travis-ci.org/ClessLi/beats-output-ding-talk-api.svg?branch=master)](https://github.com/ClessLi/beats-output-ding-talk-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/crazygreenpenguin/beats-output-http)](https://goreportcard.com/report/github.com/ClessLi/beats-output-ding-talk-api)
# beats-output-ding-talk-api
HTTP output producer for the Elastic Beats framework
beats-output-ding-talk-api. Output for the Elastic Beats platform that simply
POSTs events to an HTTP endpoint.

Compatibilities
=====
This output require v7 beat API for older version try using https://github.com/raboof/beats-output-http

Compatibility testing conducted with beats 7.8 and output version v0.0.7

Attention
=====

Not using pre-release version! It's only for tests.

Usage
=====

To add support for this output plugin to a beat, you
have to import this plugin into your main beats package,
like this:

```
package main

import (
	"os"
	_ "github.com/ClessLi/beats-output-ding-talk-api"
	"github.com/elastic/beats/v7/filebeat/cmd"
)


func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```
for filebeat it's be filebeat/main.go fo example

Plugin configuration
=====

Then configure the http output plugin in filebeat.yaml:

```yaml
output.http:
  add_fields:
    field1: 123
    field2: "fedggd"
  #add fields in out message, only if only_fields = true from v0.0.7
  url: 'https://oapi.dingtalk.com/robot/send'
  # URL for sending POST request
  max_retries: -1
  # How many retry fail send: -1=infinite, 0=no retry, default=-1
  compression: false
  # Use HTTP client compression? default=false
  keep_alive: true
  # HTTP KeepAlive using default=true
  max_idle_conns: 1
  # Controls the maximum number of idle (keep-alive) must be 1 and greater
  # Default=1
  idle_conn_timeout: 0
  # Is the maximum amount of time in seconds an idle (keep-alive)
  # connection will remain idle before closing itself.
  # Zero means no limit
  # Default=0
  response_header_timeout: 3000
  # Specifies the amount of time in milliseconds to wait for a server's response
  # headers after fully writing the request (including its body, if any).
  # This time does not include the time to read the response body.
  # default=3000ms
  api_access_token: 'test_tokenxxxxx'
  # Required, dingTalk robot webhook api access token
  # at:
  # Users assigned to view
    # at_mobiles:
    #   - '13600000000'
    # The mobile phones number of the person being @ used
    # at_user_ids:
    #   - 'clessli'
    # User IDs of the @ person
    # is_at_all: true
    # Whether @ everyone
  send_msg_type: 'text'
  # Message types pushed to API
  # Default=text
```
