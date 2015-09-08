#!/bin/bash
# This file is released into the public domain.


TOWNCRIER_BASE_URL="http://127.0.0.1:4891/receiver"

# Usage:
#
#
# post_notifications_via_curl token channel subject tag1,tag2 normal <<REQ
# Test notification
#
# Hello world!
# REQ
post_notifications_via_curl() {
  token=$1
  channel=$2
  subject=$3
  tags=$4
  priority=$5

  # --data-binary is import for us to pass in the heredoc
  # @- is for getting data from stdin
  curl -XPOST -v \
    --header "Authorization: Token token=$token" \
    --header "X-Towncrier-Subject: $subject" \
    --header "X-Towncrier-Tags: $tags" \
    --header "X-Towncrier-Priority: $priority" \
    --header "Content-Type: text/plain" \
    --data-binary @- \
    $TOWNCRIER_BASE_URL/notifications/$channel
}
