package webreceiver

import (
	"io/ioutil"
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

	notification := backend.Notification{}

	notificationContentBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	notification.Content = string(notificationContentBytes)

	notification.Subject = r.Header.Get("X-Towncrier-Subject")
	notification.Tags = strings.Split(r.Header.Get("X-Towncrier-Tags"), ",")
	notification.Channel = urlParams["channel"]
	notification.Origin = origin

	var found bool
	notification.Priority, found = backend.PriorityMap[r.Header.Get("X-Towncrier-Priority")]
	if !found {
		notification.Priority = backend.NormalPriority
	}

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
