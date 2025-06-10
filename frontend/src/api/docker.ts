import request from '@/helper/api-helper';
import { ApiListResult } from '@/types/global';

export interface ComposeCreateResult {
  log?: string;
}

export interface ComposeTestResult {
  error?: string;
  success?: boolean;
}

export interface PortHelper {
  container_port?: string;
  host_ip?: string;
  host_port?: string;
  protocol?: string;
}

export interface VolumeHelper {
  container_dir?: string;
  mode?: string;
  source_dir?: string;
  type?: string;
}

export interface ContainerOperate {
  auto_remove?: boolean;
  cmd?: string[];
  container_id?: string;
  cpu_shares?: number;
  entry_point?: string[];
  env?: string[];
  exposed_ports?: PortHelper[];
  force_pull?: boolean;
  image: string;
  ipv4?: string;
  ipv6?: string;
  labels?: string[];
  memory?: number;
  name: string;
  nano_cpus?: number;
  network?: string;
  open_stdin?: boolean;
  privileged?: boolean;
  publish_all_ports?: boolean;
  restart_policy?: string;
  tty?: boolean;
  volumes?: VolumeHelper[];
}

export interface ContainerOperation {
  names: string[];
  operation:
    | 'start'
    | 'stop'
    | 'restart'
    | 'kill'
    | 'pause'
    | 'unpause'
    | 'remove';
}

export interface ContainerResourceLimit {
  cpu?: number;
  memory?: number;
}

export interface ContainerStats {
  cache?: number;
  cpu_percent?: number;
  io_read?: number;
  io_write?: number;
  memory?: number;
  network_rx?: number;
  network_tx?: number;
  shot_time?: string;
}

export interface ContainerUpgrade {
  force_pull?: boolean;
  image: string;
  name: string;
}

export interface CreateCompose {
  compose_content?: string;
  env_content?: string;
  name?: string;
}

export interface DaemonJsonConf {
  cgroup_driver: string;
  experimental: boolean;
  fixed_cidr_v6: string;
  insecure_registries: string[];
  ip6_tables: boolean;
  ip_tables: boolean;
  ipv6: boolean;
  is_swarm: boolean;
  live_restore: boolean;
  log_max_file: string;
  log_max_size: string;
  registry_mirrors: string[];
  status: string;
  version: string;
}

export interface DockerOperation {
  operation: 'start' | 'restart' | 'stop';
}

export interface DockerStatus {
  status?: string;
}

export interface ImageBuild {
  docker_file: string;
  from: string; // edit|file
  name: string;
  tags?: string[];
}

export interface ImageLoad {
  path: string;
}

export interface ImagePull {
  image_name: string;
}

export interface ImagePush {
  name: string;
  tag_name: string;
}

export interface ImageSave {
  name: string;
  path: string;
  tag_name: string;
}

export interface ImageTag {
  source_id: string;
  target_name: string;
}

export interface Ipv6Option {
  experimental?: boolean;
  fixed_cidr_v6?: string;
  ip6_tables: boolean;
}

export interface KeyValue {
  key: string;
  value: string;
}

export interface LogOption {
  log_max_file?: string;
  log_max_size?: string;
}

export interface NetworkCreate {
  aux_address?: KeyValue[];
  aux_address_v6?: KeyValue[];
  driver: string;
  gateway?: string;
  gateway_v6?: string;
  ip_range?: string;
  ip_range_v6?: string;
  ipv4?: boolean;
  ipv6?: boolean;
  labels?: string[];
  name: string;
  options?: string[];
  subnet?: string;
  subnet_v6?: string;
}

export interface Prune {
  type: 'container' | 'image' | 'volume' | 'network' | 'buildcache';
  with_tag_all?: boolean;
}

export interface PruneResult {
  deleted_number?: number;
  space_reclaimed?: number;
}

export interface Rename {
  name: string;
  new_name: string;
}

export interface VolumeCreate {
  driver: string;
  labels?: string[];
  name: string;
  options?: string[];
}

// 查询 compose
export const queryComposeApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/compose', params);

// 查询 compose详情
export const getComposeDetailApi = (params: { name: string }) =>
  request.get('/docker/{host}/compose/detail', params);

// 更新 compose
export const updateComposeApi = (params: CreateCompose) =>
  request.put('/docker/{host}/compose', params);

// 创建 compose
export const createComposeApi = (params: CreateCompose) =>
  request.post<ComposeCreateResult>('/docker/{host}/compose', params);

// 操作 compose
export const operateComposeApi = (params: {
  name: string;
  operation: 'start' | 'stop' | 'restart' | 'up' | 'down';
}) => request.post('/docker/{host}/compose/operation', params);

// 删除 compose
export const deleteComposeApi = (params: { name: string }) =>
  request.delete('/docker/{host}/compose', params);

// 测试 compose
export const testComposeApi = (params: CreateCompose) =>
  request.post<ComposeTestResult>('/docker/{host}/compose/test', params);

// 获取 docker 配置
export const getDockerConfApi = () =>
  request.get<DaemonJsonConf>('/docker/{host}/conf');

// 更新 docker 配置 (修改字段)
export const updateDockerConfApi = (params: { key: string; value: any }) =>
  request.put('/docker/{host}/conf', params);

// 获取 docker 配置文件
export const getDockerConfRawApi = () =>
  request.get<{
    content: string;
  }>('/docker/{host}/conf/raw');

// 更新docker配置 (通过配置文件修改)
export const updateDockerConfRawApi = (params: { content: string }) =>
  request.put('/docker/{host}/conf/raw', params);

// 更新 ipv6 选项
export const updateIpv6OptionApi = (params: Ipv6Option) =>
  request.put('/docker/{host}/ipv6', params);

// 更新日志选项
export const updateLogOptionApi = (params: LogOption) =>
  request.put('/docker/{host}/log', params);

// 查询容器
export const queryContainersApi = (params: {
  info?: string;
  state: string;
  page: number;
  page_size: number;
  order_by?: string;
}) => request.get<ApiListResult<any>>('/docker/{host}/containers', params);

// 更新容器
export const updateContainerApi = (params: ContainerOperate) =>
  request.put('/docker/{host}/containers', params);

// 创建容器
export const createContainerApi = (params: ContainerOperate) =>
  request.post('/docker/{host}/containers', params);

// 获取容器详情
export const getContainerDetailApi = (id: number) =>
  request.get<ContainerOperate>('/docker/{host}/containers/detail', {
    id,
  });

// 获取容器资源限制
export const getContainerResourceLimitApi = () =>
  request.get<ContainerResourceLimit>('/docker/{host}/containers/limit');

// 清理容器日志
export const cleanContainerLogApi = (id: number) =>
  request.delete('/docker/{host}/containers/logs', { id });

// 查询容器名称
export const queryContainerNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/containers/names');

// 退出容器终端
export const quitContainerTerminalApi = (params: {
  session: string;
  data?: string;
  type: 'screen';
}) => request.post('/docker/{host}/containers/terminal/quit', params);

// 容器批量操作
export const operateContainersApi = (params: ContainerOperation) =>
  request.post('/docker/{host}/containers/operation', params);

// 容器重命名
export const renameContainerApi = (params: Rename) =>
  request.post('/docker/{host}/containers/rename', params);

// 获取容器统计
export const getContainerStatsApi = (id: number) =>
  request.get<ContainerStats>('/docker/{host}/containers/stats', { id });

// 升级容器
export const upgradeContainerApi = (params: ContainerUpgrade) =>
  request.post('/docker/{host}/containers/upgrade', params);

// 获取容器资源使用列表
export const getContainerUsagesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/containers/usages');

// 获取镜像
export const queryImagesApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/images', { ...params });

// 批量删除镜像
export const batchDeleteImagesApi = (params: {
  force: boolean;
  sources: string;
}) => request.delete('/docker/{host}/images', params);

// 构建镜像
export const buildImageApi = (params: ImageBuild) =>
  request.post('/docker/{host}/images/build', params);

// 导出镜像
export const exportImageApi = (params: ImageSave) =>
  request.post('/docker/{host}/images/export', params);

// 导入镜像
export const importImageApi = (params: ImageLoad) =>
  request.post('/docker/{host}/images/import', params);

// 查询镜像名称
export const queryImageNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/images/names');

// 拉取镜像
export const pullImageApi = (params: ImagePull) =>
  request.post('/docker/{host}/images/pull', params);

// 推送镜像
export const pushImageApi = (id: string, params: ImagePush) =>
  request.post('/docker/{host}/images/push', params, {
    params: { id },
  });

// 设置镜像标签
export const setImageTagApi = (params: ImageTag) =>
  request.put('/docker/{host}/images/tag', params);

// inspect
export const inspectApi = (params: { type: string; id: string }) =>
  request.get<any>('/docker/{host}/inspect', { ...params });

// 获取网络
export const getNetworksApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/networks', params);

// 创建网络
export const createNetworkApi = (params: NetworkCreate) =>
  request.post('/docker/{host}/networks', params);

// 批量删除网络
export const batchDeleteNetworkApi = (params: {
  force: string;
  sources: string;
}) => request.delete('/docker/{host}/networks', params);

// 查询网络名称
export const queryNetworkNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/networks/names');

// docker 操作
export const dockerOperationApi = (params: DockerOperation) =>
  request.post('/docker/{host}/operation', params);

// prune
export const pruneApi = (params: Prune) =>
  request.post<PruneResult>('/docker/{host}/prune', params);

// 获取 docker 状态
export const getDockerStatusApi = () =>
  request.get<DockerStatus>('/docker/{host}/status');

// 获取卷
export const getVolumesApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/volumes', params);

// 创建卷
export const createVolumeApi = (params: VolumeCreate) =>
  request.post('/docker/{host}/volumes', params);

// 批量删除卷
export const batchDeleteVolumeApi = (params: {
  force: string;
  sources: string;
}) => request.delete('/docker/{host}/volumes', params);

// 查询卷名称
export const queryVolumeNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/volumes/names');
