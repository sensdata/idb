package conn

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/gorilla/websocket"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type AgentWebSocketSession struct {
	Session            string
	Name               string
	SessionMessageChan chan *message.SessionMessage
	agentConn          *net.Conn
	agentSecret        string
	wsConn             *websocket.Conn
	cols               int
	rows               int
	token              string
	hostID             uint
}

func NewAgentWebSocketSession(cols, rows int, agentConn *net.Conn, wsConn *websocket.Conn, agentSecret string, token string, hostID uint) (*AgentWebSocketSession, error) {
	return &AgentWebSocketSession{
		Session:            utils.GenerateMsgId(),
		SessionMessageChan: make(chan *message.SessionMessage),
		agentConn:          agentConn,
		wsConn:             wsConn,
		agentSecret:        agentSecret,
		cols:               cols,
		rows:               rows,
		token:              token,
		hostID:             hostID,
	}, nil
}

func (s *AgentWebSocketSession) Close() {

}

func (aws *AgentWebSocketSession) Start(quitChan chan bool) {
	go aws.receiveWsMsg(quitChan)
	go aws.sendComboOutput(quitChan)
}

func (aws *AgentWebSocketSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("[xpack] A panic occurred during receive ws message, error message: %v", r)
			setQuit(exitCh)
		}
	}()
	wsConn := aws.wsConn
	defer setQuit(exitCh)

	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				// 检查是否为 CloseError
				if websocket.IsUnexpectedCloseError(err) {
					global.LOG.Info("websocket connection closed: %v", err)
					// ws断了，需要detached会话
					go aws.notifyDetach()
					return
				}
				global.LOG.Error("read message error: %v", err)
				continue
			}
			msgObj := message.WsMessage{}
			err = json.Unmarshal(wsData, &msgObj)
			if err != nil {
				global.LOG.Error("unmarshal message error: %v", err)
				continue
			}
			global.LOG.Info("receive ws msg: %v", msgObj)
			// 分发
			switch msgObj.Type {
			case message.WsMessageStart:
				aws.sendToAgent(
					aws.Session, // 初始使用aws.Session做msgId，方便channel的sessionMap查找
					message.WsMessageStart,
					message.SessionData{Type: message.SessionTypeScreen, Session: msgObj.Session, Data: msgObj.Data, Cols: aws.cols, Rows: aws.rows},
				)

			case message.WsMessageAttach:
				// 如果会话已经登记了token，说明正在被使用
				sessionToken, exist := CENTER.GetSessionToken(msgObj.Session)
				global.LOG.Info("aws token: %s \n session token: %s", aws.token, sessionToken)
				if exist && aws.token != sessionToken {
					// 别人已经在使用的提示
					errMsg := fmt.Sprintf("session %s is being used by another user", msgObj.Session)
					global.LOG.Error("%s", errMsg)
					go func() {
						msg := message.SessionMessage{
							MsgID: aws.Session,
							Type:  msgObj.Type,
							Data: message.SessionData{
								Code:    constant.CodeFailed,
								Msg:     errMsg,
								Type:    message.SessionTypeScreen,
								Session: msgObj.Session,
								Data:    "",
							},
						}
						aws.SessionMessageChan <- &msg
					}()
				} else {
					aws.sendToAgent(
						aws.Session, // 初始使用aws.Session做msgId，方便channel的sessionMap查找
						message.WsMessageAttach,
						message.SessionData{Type: message.SessionTypeScreen, Session: msgObj.Session, Data: msgObj.Data, Cols: aws.cols, Rows: aws.rows},
					)
				}

			case message.WsMessageCmd:
				aws.sendToAgent(
					utils.GenerateMsgId(),
					message.WsMessageCmd,
					message.SessionData{Type: message.SessionTypeScreen, Session: msgObj.Session, Data: msgObj.Data},
				)

			case message.WsMessageResize:
				aws.sendToAgent(
					utils.GenerateMsgId(),
					message.WsMessageResize,
					message.SessionData{Type: message.SessionTypeScreen, Session: msgObj.Session, Data: msgObj.Data, Cols: msgObj.Cols, Rows: msgObj.Rows},
				)

			case message.WsMessageHeartbeat:
				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				if err != nil {
					global.LOG.Error("sending terminal heartbeat message to webSocket failed, err: %v", err)
				}
			}

		}
	}
}

func (aws *AgentWebSocketSession) notifyDetach() {
	req := model.TerminalRequest{
		Session: aws.Session,
		Data:    "",
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		global.LOG.Error("failed to notify detach: %v", err)
		return
	}
	actionRequest := model.HostAction{
		HostID: uint(aws.hostID),
		Action: model.Action{
			Action: model.Terminal_Detach,
			Data:   data,
		},
	}
	actionResponse, err := CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return
	}
	if !actionResponse.Result {
		global.LOG.Error("failed to detach session, might already been detached")
	}
}

func (aws *AgentWebSocketSession) sendToAgent(msgId string, msgType string, data message.SessionData) error {
	// 启动监听
	msg, err := message.CreateSessionMessage(
		msgId,
		msgType,
		data,
		aws.agentSecret,
		utils.GenerateNonce(16),
	)
	if err != nil {
		global.LOG.Error("Error creating session message: %v", err)
		return err
	}

	err = message.SendSessionMessage(*aws.agentConn, msg)
	if err != nil {
		global.LOG.Error("Failed to send session message: %v", err)
		return err
	}

	return nil
}

func (aws *AgentWebSocketSession) sendComboOutput(exitCh chan bool) {
	wsConn := aws.wsConn
	defer setQuit(exitCh)

	for {
		select {
		case <-exitCh:
			return
		case response, ok := <-aws.SessionMessageChan:
			if !ok {
				global.LOG.Info("Response channel closed, exiting waitForTerminalResponse")
				return
			}
			message := message.WsMessage{
				Code:      response.Data.Code,
				Msg:       response.Data.Msg,
				Type:      string(response.Type),
				Session:   response.Data.Session,
				Data:      response.Data.Data,
				Timestamp: int(response.Timestamp),
			}
			wsData, err := json.Marshal(message)
			if err != nil {
				global.LOG.Error("encoding terminal message to json failed, err: %v", err)
				continue
			}
			err = wsConn.WriteMessage(websocket.TextMessage, wsData)
			if err != nil {
				global.LOG.Error("sending terminal message to webSocket failed, err: %v", err)
				return
			}
			global.LOG.Info("Send to ws: %s", wsData)
		}
	}
}
