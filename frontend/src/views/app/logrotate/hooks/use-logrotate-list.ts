import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { LOGROTATE_TYPE } from '@/config/enum';
import { LogrotateEntity } from '@/entity/Logrotate';
import {
  deleteLogrotateApi,
  getLogrotateListApi,
  activateLogrotateApi,
  LogrotateListParams,
} from '@/api/logrotate';
import useLoading from '@/hooks/loading';
import { useConfirm } from '@/hooks/confirm';
import useCurrentHost from '@/hooks/current-host';
import { useLogger } from '@/hooks/use-logger';
import { TABLE_CONFIG } from '../constants';

export function useLogrotateList(type: LOGROTATE_TYPE) {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { currentHostId } = useCurrentHost();

  // 日志记录
  const { logInfo } = useLogger('LogrotateList');

  // 查询参数
  const params = ref<{
    type: LOGROTATE_TYPE;
    category: string;
    page: number;
    page_size: number;
    host?: number;
  }>({
    type,
    category: '',
    page: TABLE_CONFIG.DEFAULT_PAGE,
    page_size: TABLE_CONFIG.DEFAULT_PAGE_SIZE,
    host: currentHostId.value,
  });

  // 获取配置列表
  const fetchLogrotateList = async (queryParams: LogrotateListParams) => {
    logInfo('fetchLogrotateList 被调用，参数:', queryParams);
    logInfo('参数详细信息:', {
      category: queryParams.category,
      type: queryParams.type,
      page: queryParams.page,
      page_size: queryParams.page_size,
      host: queryParams.host,
    });
    logInfo('当前 currentHostId:', currentHostId.value);
    logInfo('当前 params.value:', params.value);

    try {
      // 检查必要参数
      if (!currentHostId.value) {
        logInfo('currentHostId 为空，无法发起请求');
        Message.error('Host ID is required');
        return {
          items: [],
          total: 0,
          page: queryParams.page,
          page_size: queryParams.page_size,
        };
      }

      // 使用当前 params 中的分类，而不是传入的 queryParams
      const currentCategory = params.value.category || queryParams.category;
      logInfo('使用的分类:', currentCategory);

      // 如果 category 为空，不发起请求
      if (!currentCategory || currentCategory.trim() === '') {
        logInfo('分类为空，返回空结果');
        return {
          items: [],
          total: 0,
          page: queryParams.page,
          page_size: queryParams.page_size,
        };
      }

      // 确保 host 参数正确设置，并使用最新的分类信息
      const requestParams = {
        ...queryParams,
        category: currentCategory,
        host: currentHostId.value,
      };
      logInfo('最终请求参数:', requestParams);

      logInfo('开始请求 logrotate 列表');
      setLoading(true);
      const response = await getLogrotateListApi(requestParams);
      logInfo('logrotate 列表请求成功:', response);
      return response;
    } catch (error) {
      logInfo('logrotate 列表请求失败:', error);
      console.error('Failed to fetch logrotate list:', error);
      Message.error(t('app.logrotate.list.message.fetch_failed'));
      return {
        items: [],
        total: 0,
        page: queryParams.page,
        page_size: queryParams.page_size,
      };
    } finally {
      setLoading(false);
    }
  };

  // 删除配置
  const deleteLogrotate = async (record: LogrotateEntity): Promise<boolean> => {
    const confirmed = await confirm({
      title: t('app.logrotate.list.delete.title'),
      content: t('app.logrotate.list.delete.content', { name: record.name }),
    });

    if (!confirmed) return false;

    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.logrotate.list.message.no_host_selected'));
        return false;
      }

      await deleteLogrotateApi(
        record.type,
        record.category,
        record.name,
        currentHostId.value
      );
      Message.success(t('app.logrotate.list.message.delete_success'));
      return true;
    } catch (error) {
      console.error('Failed to delete logrotate:', error);
      Message.error(t('app.logrotate.list.message.delete_failed'));
      return false;
    }
  };

  // 激活/停用配置
  const toggleLogrotateStatus = async (
    record: LogrotateEntity,
    action: 'activate' | 'deactivate'
  ): Promise<boolean> => {
    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.logrotate.list.message.no_host_selected'));
        return false;
      }

      await activateLogrotateApi(
        record.type,
        record.category,
        record.name,
        action,
        currentHostId.value
      );

      Message.success(
        action === 'activate'
          ? t('app.logrotate.list.message.activate_success')
          : t('app.logrotate.list.message.deactivate_success')
      );
      return true;
    } catch (error) {
      console.error(`Failed to ${action} logrotate:`, error);
      Message.error(
        action === 'activate'
          ? t('app.logrotate.list.message.activate_failed')
          : t('app.logrotate.list.message.deactivate_failed')
      );
      return false;
    }
  };

  // 监听主机ID变化
  watch(
    () => currentHostId.value,
    (newHostId) => {
      params.value.host = newHostId;
    }
  );

  return {
    params,
    loading,
    fetchLogrotateList,
    deleteLogrotate,
    toggleLogrotateStatus,
  };
}
