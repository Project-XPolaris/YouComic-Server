package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

var WebsocketLogger = logrus.New().WithField("scope", "websocket")
var DefaultNotificationManager = NotificationManager{
	Conns: map[string]*NotificationConnection{},
}

type NotificationConnection struct {
	Id         string
	Connection *websocket.Conn
	Logger     *logrus.Entry
}

type NotificationManager struct {
	Conns map[string]*NotificationConnection
	sync.Mutex
}

func (m *NotificationManager) addConnection(conn *websocket.Conn) *NotificationConnection {
	m.Lock()
	defer m.Unlock()
	id := xid.New().String()
	m.Conns[id] = &NotificationConnection{
		Connection: conn,
		Logger: WebsocketLogger.WithFields(logrus.Fields{
			"id": id,
		}),
		Id: id,
	}
	return m.Conns[id]
}
func (m *NotificationManager) removeConnection(id string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Conns, id)
}
func (m *NotificationManager) sendJSONToAll(data interface{}) {
	m.Lock()
	defer m.Unlock()
	for _, notificationConnection := range m.Conns {
		err := notificationConnection.Connection.WriteJSON(data)
		if err != nil {
			notificationConnection.Logger.Error(err)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WShandler(context *gin.Context) {
	c, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		WebsocketLogger.Error(err)
		return
	}
	notifier := DefaultNotificationManager.addConnection(c)
	notifier.Logger.Info("notification added")
	defer func() {
		DefaultNotificationManager.removeConnection(notifier.Id)
		c.Close()
		notifier.Logger.Info("notification disconnected")
	}()
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 1005, 1000) {
				notifier.Logger.Error(err)
			}
			break
		}
	}
}
