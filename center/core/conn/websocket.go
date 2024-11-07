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
	AuthMode   string `json:"auth_mode"`
	Password   string `json:"password"`
	PrivateKey []byte `json:"private_key"`
	PassPhrase []byte `json:"pass_phrase"`

	Client     *gossh.Client  `json:"client"`
	Session    *gossh.Session `json:"session"`
	LastResult string         `json:"last_result"`
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
		return errors.Wrap(err, "failed to upgrade to WebSocket")
	}
	defer wsConn.Close()

	global.LOG.Info("upgrade successful")

	hostID, err := strconv.Atoi(c.Query("host_id"))
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "invalid param id in request")
	}
	cols, err := strconv.Atoi(c.DefaultQuery("cols", "80"))
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "invalid param cols in request")
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "40"))
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "invalid param rows in request")
	}
	host, err := HostRepo.Get(HostRepo.WithByID((uint(hostID))))
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "load host info by id failed")
	}

	global.LOG.Info("private key content: \n %s", host.PrivateKey)

	// 建立新的ssh连接
	var connInfo SshConn
	_ = copier.Copy(&connInfo, &host)

	// Decode private key after retrieving
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(host.PrivateKey)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "failed to decode private key")
	}
	connInfo.PrivateKey = decodedPrivateKey
	if len(host.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(host.PassPhrase)
	}

	client, err := connInfo.NewSshClient()
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "failed to set up the connection. Please check the host information")
	}
	defer client.Close()

	sws, err := NewSshWebSocketSession(cols, rows, true, connInfo.Client, wsConn)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "failed to create SSH WebSocket session")
	}
	defer sws.Close()

	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	<-quitChan

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
	global.LOG.Info("authmode: %s, %s", c.AuthMode, c.Password)
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
