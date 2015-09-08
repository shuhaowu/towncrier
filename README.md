Towncrier
=========

Send notifications/alerts via HTTP. Supports many backends for sending notifications (email, phone, whatever you want), but all clients needs to do is to send it via HTTP (one of the lowest common demoninator).

Other "frontends" (UDP, whatever) maybe possible, but not yet implemented.

Note: if you're viewing this on Github, the official repository is available at https://gitlab.com/shuhao/towncrier. The Github repository serves as a mirror.

Detailed documentations available here: WIP

Current Features
----------------

- Send email notifications via HTTP
- Send notifications to different subscribers using channels
- Batch notifications by cron expressions associated with channels
- Supports pluggable backends.

Example use case
----------------

- Centralized location to send emails (only need to run 1 email server somewhere).
- Aggregate cron/alert/status emails and send it to people.
- Enable easy notification sending from software running on platforms/computers you do not control (Windows, OS X).
- A dashboard to past notifications in a central place.

Sending notifications to Towncrier
----------------------------------

Sending notifications for towncriers to send out is designed to be the easiest thing, available on all platforms. Example scripts in different environments are provided under the directory `example_clients`. However, if it is missing your platform, the API is as follows:

```
POST /<PathPrefix>/notifications/<channel> HTTP/1.1
Authorization: Token token=<your defined token>
X-Towncrier-Subject: subject line
X-Towncrier-Tags: tag1,tag2
X-Towncrier-Priority: normal

Notification content goes here.

All notification content will be sent as plain-text.
```

Detailed documentations available here: WIP

Development Setup
-----------------

1. Install [`godep`](https://github.com/tools/godep) and [`goose`](https://bitbucket.org/liamstask/goose).
2. Clone the repository into `$GOPATH/src/gitlab.com/shuhao/towncrier`.
3. `godep go test ./...`
4. `script/devserver`

License
-------

AGPLv3
