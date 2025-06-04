export interface ServiceEntity {
  id?: number;
  name: string;
  type: string;
  source: string;
  size: number;
  mod_time: string;
  linked?: boolean; // 服务是否已激活
  status?: string; // 服务运行状态
  description?: string;
  content?: string;
  commit?: string;
  version?: string;
  extension?: string; // 文件扩展名
}

export interface ServiceLogEntity {
  timestamp: string;
  level: string;
  message: string;
}

export interface ServiceHistoryEntity {
  commit: string;
  author: string;
  date: string;
  message: string;
  changes: number;
}
