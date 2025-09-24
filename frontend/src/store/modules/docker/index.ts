import { defineStore } from 'pinia';
import { getDockerInstallStatusApi } from '@/api/docker';

export type DockerInstallStatus =
  | 'unknown'
  | 'installed'
  | 'not installed'
  | 'checking';

interface HostStatus {
  status: DockerInstallStatus;
  lastChecked: number; // epoch ms
}

interface State {
  byHost: Record<string, HostStatus>;
  currentHostId: string | null;
}

const DEFAULT_TTL = 60_000; // 60s

const useDockerStatusStore = defineStore('dockerStatus', {
  state: (): State => ({
    byHost: {},
    currentHostId: null,
  }),

  getters: {
    currentStatus(state): DockerInstallStatus {
      const key = state.currentHostId ?? '';
      return state.byHost[key]?.status ?? 'unknown';
    },
    isInstalled(state): boolean {
      const key = state.currentHostId ?? '';
      return (state.byHost[key]?.status ?? 'unknown') === 'installed';
    },
    isNotInstalled(state): boolean {
      const key = state.currentHostId ?? '';
      return (state.byHost[key]?.status ?? 'unknown') === 'not installed';
    },
    isFresh(state) {
      return (hostId?: string | null, ttl: number = DEFAULT_TTL) => {
        const key = hostId ?? state.currentHostId ?? '';
        const last = state.byHost[key]?.lastChecked ?? 0;
        return Date.now() - last < ttl;
      };
    },
  },

  actions: {
    setCurrentHost(hostId: string | null) {
      this.currentHostId = hostId ?? null;
      const key = hostId ?? '';
      if (!this.byHost[key]) {
        this.byHost[key] = { status: 'unknown', lastChecked: 0 };
      }
    },

    setStatus(status: DockerInstallStatus, hostId?: string | null) {
      const key = hostId ?? this.currentHostId ?? '';
      if (!this.byHost[key])
        this.byHost[key] = { status: 'unknown', lastChecked: 0 };
      this.byHost[key].status = status;
      this.byHost[key].lastChecked = Date.now();
    },

    async refresh(hostId?: string | null): Promise<DockerInstallStatus> {
      const key = hostId ?? this.currentHostId ?? '';
      try {
        const res = await getDockerInstallStatusApi();
        const s = (res?.status as DockerInstallStatus) || 'unknown';
        this.setStatus(s, key);
        return this.byHost[key].status;
      } catch {
        // 检测失败时设置为 unknown，避免错误显示安装提示
        this.setStatus('unknown', key);
        return this.byHost[key]?.status ?? 'unknown';
      }
    },

    async ensureChecked(
      hostId?: string | null,
      maxAgeMs: number = DEFAULT_TTL
    ): Promise<DockerInstallStatus> {
      const key = hostId ?? this.currentHostId ?? '';
      const status = this.byHost[key]?.status ?? 'unknown';
      if (status === 'unknown' || !this.isFresh(key, maxAgeMs)) {
        return this.refresh(key);
      }
      return status;
    },
  },
});

export default useDockerStatusStore;
