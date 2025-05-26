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
    | 'resume'
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
  cgroup_driver?: string;
  experimental?: boolean;
  fixed_cidr_v6?: string;
  insecure_registries?: string[];
  ip6_tables?: boolean;
  ip_tables?: boolean;
  ipv6?: boolean;
  is_swarm?: boolean;
  live_restore?: boolean;
  log_max_file?: string;
  log_max_size?: string;
  registry_mirrors?: string[];
  status?: string;
  version?: string;
}

export interface DaemonJsonUpdateByFile {
  file?: string;
}

export interface DockerOperation {
  operation: 'start' | 'restart' | 'stop';
}

export interface DockerStatus {
  status?: string;
}

export interface ImageBuild {
  docker_file: string;
  from: string;
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

// 更新 compose
export const updateComposeApi = (data: CreateCompose) =>
  request.put('/docker/{host}/compose', data);

// 创建 compose
export const createComposeApi = (data: CreateCompose) =>
  request.post<ComposeCreateResult>('/docker/{host}/compose', data);

// 操作 compose
export const operateComposeApi = (data: {
  name: string;
  operation: 'start' | 'stop' | 'down';
}) => request.post('/docker/{host}/compose/operation', data);

// 测试 compose
export const testComposeApi = (data: CreateCompose) =>
  request.post<ComposeTestResult>('/docker/{host}/compose/test', data);

// 获取 docker 配置
export const getDockerConfApi = () =>
  request.get<DaemonJsonConf>('/docker/{host}/conf');

// 更新 docker 配置
export const updateDockerConfApi = (data: DaemonJsonUpdateByFile) =>
  request.put('/docker/{host}/conf', data);

// 查询容器
export const queryContainersApi = (params: {
  info?: string;
  state: string;
  page: number;
  page_size: number;
  order_by?: string;
}) => request.get<ApiListResult<any>>('/docker/{host}/containers', params);

// 更新容器
export const updateContainerApi = (data: ContainerOperate) =>
  request.put('/docker/{host}/containers', data);

// 创建容器
export const createContainerApi = (data: ContainerOperate) =>
  request.post('/docker/{host}/containers', data);

// 获取容器详情
export const getContainerDetailApi = (id: number) =>
  request.get<ContainerOperate>('/docker/{host}/containers/detail', {
    id,
  });

// 获取容器资源限制
export const getContainerResourceLimitApi = () =>
  request.get<ContainerResourceLimit>('/docker/{host}/containers/limit');

// 获取容器日志
export const getContainerLogApi = (id: number) =>
  request.get<any>('/docker/{host}/containers/log', { id });

// 清理容器日志
export const cleanContainerLogApi = (id: number) =>
  request.delete('/docker/{host}/containers/log', { id });

// 查询容器名称
export const queryContainerNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/containers/names');

// 容器批量操作
export const operateContainersApi = (data: ContainerOperation) =>
  request.post('/docker/{host}/containers/operatetion', data);

// 容器重命名
export const renameContainerApi = (data: Rename) =>
  request.post('/docker/{host}/containers/rename', data);

// 获取容器统计
export const getContainerStatsApi = (id: number) =>
  request.get<ContainerStats>('/docker/{host}/containers/stats', { id });

// 升级容器
export const upgradeContainerApi = (data: ContainerUpgrade) =>
  request.post('/docker/{host}/containers/upgrade', data);

// 获取容器资源使用列表
export const getContainerUsagesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/containers/usages');

// 获取镜像
export const getImagesApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/images', { ...params });

// 批量删除镜像
export const batchDeleteImagesApi = (params: {
  force: string;
  sources: string;
}) => request.delete('/docker/{host}/images', params);

// 构建镜像
export const buildImageApi = (data: ImageBuild) =>
  request.post('/docker/{host}/images/build', data);

// 导出镜像
export const exportImageApi = (data: ImageSave) =>
  request.post('/docker/{host}/images/export', data);

// 导入镜像
export const importImageApi = (data: ImageLoad) =>
  request.post('/docker/{host}/images/import', data);

// 查询镜像名称
export const queryImageNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/images/names');

// 拉取镜像
export const pullImageApi = (data: ImagePull) =>
  request.post('/docker/{host}/images/pull', data);

// 推送镜像
export const pushImageApi = (id: string, data: ImagePush) =>
  request.post('/docker/{host}/images/push', data, {
    params: { id },
  });

// 设置镜像标签
export const setImageTagApi = (data: ImageTag) =>
  request.put('/docker/{host}/images/tag', data);

// inspect
export const inspectApi = (params: { type: string; id: string }) =>
  request.get<any>('/docker/{host}/inspect', { ...params });

// 更新 ipv6 选项
export const updateIpv6OptionApi = (data: Ipv6Option) =>
  request.put('/docker/{host}/ipv6', data);

// 更新日志选项
export const updateLogOptionApi = (data: LogOption) =>
  request.put('/docker/{host}/log', data);

// 获取网络
export const getNetworksApi = (params: {
  info?: string;
  page: number;
  page_size: number;
}) => request.get<ApiListResult<any>>('/docker/{host}/networks', params);

// 创建网络
export const createNetworkApi = (data: NetworkCreate) =>
  request.post('/docker/{host}/networks', data);

// 批量删除网络
export const batchDeleteNetworkApi = (params: {
  force: string;
  sources: string;
}) => request.delete('/docker/{host}/networks', params);

// 查询网络名称
export const queryNetworkNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/networks/names');

// docker 操作
export const dockerOperationApi = (data: DockerOperation) =>
  request.post('/docker/{host}/operation', data);

// prune
export const pruneApi = (data: Prune) =>
  request.post<PruneResult>('/docker/{host}/prune', data);

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
export const createVolumeApi = (data: VolumeCreate) =>
  request.post('/docker/{host}/volumes', data);

// 批量删除卷
export const batchDeleteVolumeApi = (params: {
  force: string;
  sources: string;
}) => request.delete('/docker/{host}/volumes', params);

// 查询卷名称
export const queryVolumeNamesApi = () =>
  request.get<ApiListResult<any>>('/docker/{host}/volumes/names');
