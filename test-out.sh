#!/usr/bin/env sh

go build ./cmd/out

echo \
'{
  "source": {
    "defectdojo_url": "https://something",
    "api_key": "key",
    "product_name": "app",
    "debug": true
  },
  "params": {
    "report_type": "ZAP Scan",
    "path_to_report": "something",
    "close_engagement": true
  }
}' | ./out

rm out
