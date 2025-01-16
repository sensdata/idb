export enum MsgType {
  Heartbeat = 'heartbeat',
  Cmd = 'cmd',
  Attach = 'attach',
  Start = 'start',
  Resize = 'resize',
}

export interface SendMsgDo {
  type: MsgType;
  data?: string;
  session?: string;
  cols?: number;
  rows?: number;
  timestamp?: number;
}

export interface ReceiveMsgDo extends SendMsgDo {
  code: number;
  msg: string;
}
