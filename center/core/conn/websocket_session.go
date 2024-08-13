package conn

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
	"golang.org/x/crypto/ssh"
)

type SshWebSocketSession struct {
	stdinPipe       io.WriteCloser
	comboOutput     *utils.SafeBuffer
	logBuff         *utils.SafeBuffer
	inputFilterBuff *utils.SafeBuffer
	session         *ssh.Session
	wsConn          *websocket.Conn
	isAdmin         bool
	IsFlagged       bool
}

func NewSshWebSocketSession(cols, rows int, isAdmin bool, sshClient *ssh.Client, wsConn *websocket.Conn) (*SshWebSocketSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(utils.SafeBuffer)
	logBuf := new(utils.SafeBuffer)
	inputBuf := new(utils.SafeBuffer)
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	return &SshWebSocketSession{
		stdinPipe:       stdinP,
		comboOutput:     comboWriter,
		logBuff:         logBuf,
		inputFilterBuff: inputBuf,
		session:         sshSession,
		wsConn:          wsConn,
		isAdmin:         isAdmin,
		IsFlagged:       false,
	}, nil
}

func (s *SshWebSocketSession) Close() {
	if s.session != nil {
		s.session.Close()
	}
	if s.logBuff != nil {
		s.logBuff = nil
	}
	if s.comboOutput != nil {
		s.comboOutput = nil
	}
}

func (sws *SshWebSocketSession) Start(quitChan chan bool) {
	go sws.receiveWsMsg(quitChan)
	go sws.sendComboOutput(quitChan)
}

func (sws *SshWebSocketSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("[xpack] A panic occurred during receive ws message, error message: %v", r)
		}
	}()
	wsConn := sws.wsConn
	defer setQuit(exitCh)
	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				return
			}
			msgObj := message.WsMessage{}
			_ = json.Unmarshal(wsData, &msgObj)
			switch msgObj.Type {
			case message.WsMessageResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.session.WindowChange(msgObj.Rows, msgObj.Cols); err != nil {
						global.LOG.Error("ssh pty change windows size failed, err: %v", err)
					}
				}
			case message.WsMessageCmd:
				decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
				if err != nil {
					global.LOG.Error("websock cmd string base64 decoding failed, err: %v", err)
				}
				sws.socketInputToSshPipe(decodeBytes)
			case message.WsMessageHeartbeat:
				// 接收到心跳包后将心跳包原样返回，可以用于网络延迟检测等情况
				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				if err != nil {
					global.LOG.Error("ssh sending heartbeat to webSocket failed, err: %v", err)
				}
			}
		}
	}
}

func (sws *SshWebSocketSession) socketInputToSshPipe(cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		global.LOG.Error("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

func (sws *SshWebSocketSession) sendComboOutput(exitCh chan bool) {
	wsConn := sws.wsConn
	defer setQuit(exitCh)

	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if sws.comboOutput == nil {
				return
			}
			bs := sws.comboOutput.Bytes()
			if len(bs) > 0 {
				wsData, err := json.Marshal(message.WsMessage{
					Type: message.WsMessageCmd,
					Data: base64.StdEncoding.EncodeToString(bs),
				})
				if err != nil {
					global.LOG.Error("encoding combo output to json failed, err: %v", err)
					continue
				}
				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				if err != nil {
					global.LOG.Error("ssh sending combo output to webSocket failed, err: %v", err)
				}
				_, err = sws.logBuff.Write(bs)
				if err != nil {
					global.LOG.Error("combo output to log buffer failed, err: %v", err)
				}
				sws.comboOutput.Buffer.Reset()
			}
			if string(bs) == string([]byte{13, 10, 108, 111, 103, 111, 117, 116, 13, 10}) {
				sws.Close()
				return
			}

		case <-exitCh:
			return
		}
	}
}

func (sws *SshWebSocketSession) Wait(quitChan chan bool) {
	if err := sws.session.Wait(); err != nil {
		setQuit(quitChan)
	}
}

func setQuit(ch chan bool) {
	ch <- true
}
