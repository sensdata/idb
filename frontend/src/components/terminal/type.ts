export enum MsgType {
  Heartbeat = 'heartbeat',
  Command = 'command',
  Attach = 'attach',
  Start = 'start',
}

export interface MsgDo {
  type: MsgType;
  data?: string;
  session?: string;
  cols?: number;
  rows?: number;
  timestamp?: number;
}
