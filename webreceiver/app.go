package webreceiver

import (
	"encoding/json"
	"net/http"
	"strings"

	"gitlab.com/shuhao/towncrier/backend"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// The webreceiver app simply receives and queues the notifications.
// Displaying the notification is the job of the webfeed package

var realLogger = logrus.New()
var logger = realLogger.WithField("component", "webreceiver")

type NotificationData struct {
	Subject  string
	Content  string
	Tags     []string
	Priority string
}

func (n NotificationData) ToNotification(channel, origin string) backend.Notification {
	priority, found := backend.PriorityMap[n.Priority]
	if !found {
		priority = backend.NormalPriority
	}

	return backend.Notification{
		Subject:  n.Subject,
		Content:  n.Content,
		Tags:     n.Tags,
		Priority: priority,
		Channel:  channel,
		Origin:   origin,
	}
}

type App struct {
	router  *mux.Router
	config  ReceiverConfig
	backend backend.NotificationBackend
}

func NewApp(be backend.NotificationBackend, config ReceiverConfig) *App {
	app := &App{
		config:  config,
		backend: be,
	}

	app.router = mux.NewRouter()
	subrouter := app.router.PathPrefix(app.config.PathPrefix)
	subrouter.Methods("POST").Path("/notification/{channel}").HandlerFunc(app.PostNotificationHandler)

	return app
}

func (a *App) isAuthenticated(r *http.Request) (authenticated bool, token string, origin string) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false, "", ""
	}

	authArray := strings.Split(auth, " ")
	if len(authArray) != 2 || authArray[0] != "Token" {
		return false, "", ""
	}

	tokenArray := strings.Split(authArray[1], "=")
	if len(tokenArray) != 2 || tokenArray[0] != "token" {
		return false, "", ""
	}

	token = tokenArray[1]
	origin, found := a.config.Tokens[token]
	return found, token, origin
}

func (a *App) PostNotificationHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, _, origin := a.isAuthenticated(r)
	if !authenticated {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	urlParams := mux.Vars(r)

	notificationData := NotificationData{}
	err := json.NewDecoder(r.Body).Decode(&notificationData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	notification := notificationData.ToNotification(urlParams["channel"], origin)
	err = a.backend.QueueNotification(notification)
	if err != nil {
		if _, ok := err.(backend.ChannelNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger.WithField("error", err).Error("failed to queue notification")
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
