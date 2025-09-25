import { Message } from '@arco-design/web-vue';
import useDockerStatusStore from '@/store/modules/docker';

/**
 * 统一的错误提示封装：
 * - 默认会在疑似 Docker 相关错误时，检查是否未安装 Docker；
 * - 若确实未安装，则不弹错误提示（交给页面上的 Docker 安装组件引导）。
 * - 返回值：true 表示已显示错误，false 表示被抑制（未显示）。
 */
export async function showErrorWithDockerCheck(
  message: string,
  _error?: unknown,
  options?: {
    ensureFreshMaxAgeMs?: number;
  }
): Promise<boolean> {
  const { ensureFreshMaxAgeMs = 60_000 } = options || {};

  const dockerStore = useDockerStatusStore();
  await dockerStore.ensureChecked(null, ensureFreshMaxAgeMs);

  if (dockerStore.isNotInstalled) {
    return false;
  }

  Message.error(message);
  return true;
}
