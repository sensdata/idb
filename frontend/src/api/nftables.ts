import request from '@/helper/api-helper';
import { ApiListResult } from '@/types/global';

// 类型定义
export type ConfigType = 'global' | 'local';

// 接口定义
export interface ProcessStatus {
  pid: number;
  port: number;
  process: string;
  addresses: string[];
}

export interface ProcessStatusResponse {
  total: number;
  items: ProcessStatus[];
}

export interface NftablesStatus {
  status: string; // installed/not installed
  active: string; // 当前激活的防火墙系统
}

// 配置文件详情接口
export interface ConfigFileDetail {
  source: string;
  name: string;
  extension: string;
  content: string;
  size: number;
  mod_time: string;
  linked: boolean;
}

// API 参数接口
export interface ActivateConfigApiParams {
  type: ConfigType;
  category: string;
  name: string;
  action: 'activate' | 'deactivate';
}

export interface NftablesRawConfig {
  content: string;
}

// 端口管理相关接口
// 规则项类型
export interface RuleItem {
  type: 'default' | 'rate_limit' | 'concurrent_limit';
  rate?: string; // 速率限制，如 "100/second"
  count?: number; // 并发限制数量
  action: 'accept' | 'drop' | 'reject';
}

// 基础端口规则（用于可视化配置）
export interface PortRule {
  port: number | number[]; // 支持单个端口或端口数组
  protocol?: 'tcp' | 'udp' | 'both'; // 可选，从配置中解析，无法确定时为空
  action?: 'accept' | 'drop' | 'reject'; // 可选，从配置中解析，无法确定时为空
  description?: string;
  source?: string; // 源IP或网段
  destination?: string; // 目标IP或网段
  // 高级规则
  rules?: RuleItem[]; // 高级规则列表
}

export interface PortRuleSet {
  port: number;
  rules: RuleItem[];
}

export interface SetPortRuleApiParams {
  port: number;
  rules: RuleItem[];
}

// 配置激活 API
export function activateConfigApi(data: ActivateConfigApiParams) {
  return request.post('/nftables/{host}/activate', data);
}

// 系统管理 API
export function installApi() {
  return request.post('/nftables/{host}/install');
}

export function switchFirewallApi(data: { option: 'nftables' | 'iptables' }) {
  return request.post('/nftables/{host}/switch/to', data);
}

// 进程状态 API
export function getProcessStatusApi() {
  return request.get<ProcessStatusResponse>('/nftables/{host}/process');
}

// 获取防火墙状态 API
export function getFirewallStatusApi() {
  return request.get<NftablesStatus>('/nftables/{host}/status');
}

// 获取nftables raw配置 API
export function getNftablesRawConfigApi() {
  return request.get<NftablesRawConfig>('/nftables/{host}/conf/raw');
}

// 更新nftables raw配置 API
export function updateNftablesRawConfigApi(data: NftablesRawConfig) {
  return request.put('/nftables/{host}/conf/raw', data);
}

// 端口管理 API
export function getPortRulesApi() {
  return request.get<ApiListResult<PortRuleSet>>('/nftables/{host}/port');
}

export function setPortRulesApi(data: SetPortRuleApiParams) {
  return request.post('/nftables/{host}/port/rules', data);
}

export function deletePortRulesApi(params: { port: number }) {
  return request.delete('/nftables/{host}/port/rules', params);
}

// IP黑名单管理相关接口
export interface IPBlacklistRequest {
  ip: string;
}

export interface DeleteIPBlacklistRequest {
  ip: string;
}

// IP黑名单 API
export function getIPBlacklistApi() {
  return request.get<ApiListResult<string>>('/nftables/{host}/ip/blacklist');
}

export function addIPBlacklistApi(data: IPBlacklistRequest) {
  return request.post('/nftables/{host}/ip/blacklist', data);
}

export function deleteIPBlacklistApi(params: DeleteIPBlacklistRequest) {
  return request.delete('/nftables/{host}/ip/blacklist', params);
}

// Ping 管理相关接口
export interface PingStatus {
  allowed: boolean;
}

export interface SetPingStatusRequest {
  allowed: boolean;
}

// 获取 ping 状态
export function getPingStatusApi() {
  return request.get<PingStatus>('/nftables/{host}/ping');
}

// 设置 ping 状态
export function setPingStatusApi(data: SetPingStatusRequest) {
  return request.post('/nftables/{host}/ping', data);
}
