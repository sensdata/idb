package conn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/message"
	gossh "golang.org/x/crypto/ssh"
)

type WebSocketService struct{}

type IWebSocketService interface {
	HandleTerminal(c *gin.Context) error
}

type SshConn struct {
	User       string `json:"user"`
	Addr       string `json:"addr"`
	Port       int    `json:"port"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	PrivateKey []byte `json:"privateKey"`
	PassPhrase []byte `json:"passPhrase"`

	Client     *gossh.Client  `json:"client"`
	Session    *gossh.Session `json:"session"`
	LastResult string         `json:"lastResult"`
}

func NewIWebSocketService() IWebSocketService {
	return &WebSocketService{}
}

func (s *WebSocketService) HandleTerminal(c *gin.Context) error {
	global.LOG.Info("handle terminal begin")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Error("Failed to upgrade to WebSocket: %v\n", err)
		return err
	}
	defer wsConn.Close()

	global.LOG.Info("upgrade")

	id, err := strconv.Atoi(c.Query("id"))
	if wsHandleError(wsConn, errors.WithMessage(err, "invalid param id in request")) {
		return err
	}
	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if wsHandleError(wsConn, errors.WithMessage(err, "invalid param cols in request")) {
		return err
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if wsHandleError(wsConn, errors.WithMessage(err, "invalid param rows in request")) {
		return err
	}
	host, err := HostRepo.Get(HostRepo.WithByID((uint(id))))
	if wsHandleError(wsConn, errors.WithMessage(err, "load host info by id failed")) {
		return err
	}

	// 建立新的ssh连接
	var connInfo SshConn
	_ = copier.Copy(&connInfo, &host)
	connInfo.PrivateKey = []byte(host.PrivateKey)
	if len(host.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(host.PassPhrase)
	}

	client, err := connInfo.NewSshClient()
	if wsHandleError(wsConn, errors.WithMessage(err, "failed to set up the connection. Please check the host information")) {
		return err
	}
	defer client.Close()

	sws, err := NewSshWebSocketSession(cols, rows, true, connInfo.Client, wsConn)
	if wsHandleError(wsConn, err) {
		return err
	}
	defer sws.Close()

	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	<-quitChan

	if wsHandleError(wsConn, err) {
		return err
	}

	global.LOG.Info("handle terminal end")
	return nil
}

func (c *SshConn) NewSshClient() (*SshConn, error) {
	proto := "tcp"
	addr := c.Addr
	if strings.Contains(c.Addr, ":") {
		addr = fmt.Sprintf("[%s]", c.Addr)
		proto = "tcp6"
	}
	dialAddr := fmt.Sprintf("%s:%d", addr, c.Port)

	global.LOG.Info("try connect to host ssh: %s", dialAddr)

	//connection config
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.User = c.User
	if c.AuthMode == "password" {
		config.Auth = []gossh.AuthMethod{gossh.Password(c.Password)}
	} else {
		signer, err := makePrivateKeySigner(c.PrivateKey, c.PassPhrase)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", c.Addr, err)
			return nil, err
		}
		config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	}
	config.Timeout = 5 * time.Second
	config.HostKeyCallback = gossh.InsecureIgnoreHostKey()

	client, err := gossh.Dial(proto, dialAddr, config)
	if nil != err {
		global.LOG.Error("Failed to create ssh connection to host %s, %v", c.Addr, err)
		return c, err
	}
	c.Client = client

	global.LOG.Info("SSH connection to %s created", addr)
	return c, nil
}

func (c *SshConn) Close() {
	_ = c.Client.Close()
}

func wsHandleError(ws *websocket.Conn, err error) bool {
	if err != nil {
		global.LOG.Error("handler ws faled:, err: %v", err)
		dt := time.Now().Add(time.Second)
		if ctlerr := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); ctlerr != nil {
			wsData, err := json.Marshal(message.WsMessage{
				Type: message.WsMessageCmd,
				Data: base64.StdEncoding.EncodeToString([]byte(err.Error())),
			})
			if err != nil {
				_ = ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"cmd\",\"data\":\"failed to encoding to json\"}"))
			} else {
				_ = ws.WriteMessage(websocket.TextMessage, wsData)
			}
		}
		return true
	}
	return false
}
