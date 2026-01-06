package conn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
	gossh "golang.org/x/crypto/ssh"
)

type WebSocketService struct{}

type IWebSocketService interface {
	HandleSshTerminal(c *gin.Context) error
	HandleAgentTerminal(c *gin.Context) error
}

type SshConn struct {
	User       string `json:"user"`
	Addr       string `json:"addr"`
	Port       int    `json:"port"`
	AuthMode   string `json:"auth_mode"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`

	Client     *gossh.Client  `json:"client"`
	Session    *gossh.Session `json:"session"`
	LastResult string         `json:"last_result"`
}

func NewWebSocketService() IWebSocketService {
	return &WebSocketService{}
}

func (s *WebSocketService) HandleSshTerminal(c *gin.Context) error {
	global.LOG.Info("handle ssh terminal begin")
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

	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "invalid param host in request")
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

	// 建立新的ssh连接
	var connInfo SshConn
	_ = copier.Copy(&connInfo, &host)

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

	global.LOG.Info("handle ssh terminal end")
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
		// 读取宿主机文件, 需要利用agent连接来读取文件内容
		privateKey, err := getPrivateKey(c.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			return nil, errors.New(constant.ErrFileRead)
		}
		passPhrase := []byte(c.PassPhrase)

		signer, err := makePrivateKeySigner([]byte(privateKey.Content), passPhrase)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", c.Addr, err)
			return nil, fmt.Errorf("failed to config private key to host %s, %v", c.Addr, err)
		}
		config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	}
	config.Timeout = 5 * time.Second

	hostPort := net.JoinHostPort(c.Addr, strconv.Itoa(c.Port))
	cb, err := utils.NewHostKeyCallback(
		filepath.Join(constant.CenterBinDir, ".ssh", "known_hosts"),
		hostPort,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key callback: %w", err)
	}
	config.HostKeyCallback = cb

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
		global.LOG.Error("handler ws failed:, err: %v", err)
		dt := time.Now().Add(time.Second)
		if ctlerr := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); ctlerr != nil {
			wsData, err := json.Marshal(message.WsMessage{
				Code: constant.CodeFailed,
				Msg:  base64.StdEncoding.EncodeToString([]byte(err.Error())),
				Type: message.WsMessageCmd,
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

func (s *WebSocketService) HandleAgentTerminal(c *gin.Context) error {
	global.LOG.Info("handle agent terminal begin")
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

	token, _ := c.Cookie("idb-token")

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
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "invalid param host in request")
	}
	// 会话类型，未传默认为screen
	var sessionType message.SessionType
	st := c.Query("type")
	if st == "" {
		sessionType = message.SessionTypeScreen
	} else {
		sessionType = message.SessionType(st)
	}

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(uint(hostID)))
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "no host found")
	}
	agentConn, err := CENTER.GetAgentConn(&host)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "agent disconected")
	}

	aws, err := NewAgentWebSocketSession(cols, rows, agentConn, wsConn, host.AgentKey, token, uint(hostID), sessionType)
	if err != nil {
		wsHandleError(wsConn, err)
		return errors.Wrap(err, "failed to create Agent WebSocket session")
	}
	defer aws.Close()

	quitChan := make(chan bool, 3)
	// 将 aws 记录到center中
	CENTER.RegisterAgentSession(aws)
	CENTER.RegisterSessionToken(aws.Session, token)
	aws.Start(quitChan)
	// 等待quitChan
	<-quitChan
	// 将 aws 从center中清除
	CENTER.UnregisterSessionToken(aws.Session)
	CENTER.UnregisterAgentSession(aws.Session)

	global.LOG.Info("handle agent terminal end")
	return nil
}
