import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { SERVICE_TYPE, SERVICE_ACTION, SERVICE_OPERATION } from '@/config/enum';
import { ServiceEntity } from '@/entity/Service';
import {
  deleteServiceApi,
  getServiceListApi,
  serviceActivateApi,
  serviceOperateApi,
  ServiceListApiParams,
} from '@/api/service';
import useLoading from '@/hooks/loading';
import { useConfirm } from '@/hooks/confirm';
import useCurrentHost from '@/hooks/current-host';

export function useServiceList(type: SERVICE_TYPE) {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { currentHostId } = useCurrentHost();

  // 查询参数
  const params = ref<{
    type: SERVICE_TYPE;
    category: string;
    page: number;
    page_size: number;
    host?: number;
  }>({
    type,
    category: '',
    page: 1,
    page_size: 20,
    host: currentHostId.value,
  });

  // 获取服务列表
  const fetchServiceList = async (queryParams: ServiceListApiParams) => {
    try {
      // 检查必要参数
      if (!currentHostId.value) {
        Message.error('Host ID is required');
        return {
          items: [],
          total: 0,
          page: queryParams.page || 1,
          page_size: queryParams.page_size || 20,
        };
      }

      // 使用当前 params 中的分类，而不是传入的 queryParams
      const currentCategory = params.value.category || queryParams.category;

      // 如果 category 为空，不发起请求
      if (!currentCategory || currentCategory.trim() === '') {
        return {
          items: [],
          total: 0,
          page: queryParams.page || 1,
          page_size: queryParams.page_size || 20,
        };
      }

      // 确保 host 参数正确设置，并使用最新的分类信息
      const requestParams = {
        ...queryParams,
        category: currentCategory,
        host: currentHostId.value,
      };

      setLoading(true);
      const response = await getServiceListApi(requestParams);
      return response;
    } catch (error) {
      console.error('Failed to fetch service list:', error);
      Message.error(t('app.service.list.error.fetch'));
      return {
        items: [],
        total: 0,
        page: queryParams.page || 1,
        page_size: queryParams.page_size || 20,
      };
    } finally {
      setLoading(false);
    }
  };

  // 删除服务
  const deleteService = async (record: ServiceEntity): Promise<boolean> => {
    const confirmed = await confirm(
      t('app.service.list.confirm.delete', { name: record.name })
    );

    if (!confirmed) return false;

    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.service.list.error.no_host'));
        return false;
      }

      // 使用当前选中的分类
      if (!params.value.category || params.value.category.trim() === '') {
        Message.error(t('app.service.list.error.invalid_category'));
        return false;
      }

      await deleteServiceApi({
        type,
        category: params.value.category,
        name: record.name,
      });
      Message.success(t('app.service.list.success.delete'));
      return true;
    } catch (error) {
      console.error('Failed to delete service:', error);
      Message.error(t('app.service.list.error.delete'));
      return false;
    }
  };

  // 激活/停用服务
  const toggleServiceStatus = async (
    record: ServiceEntity,
    action: SERVICE_ACTION
  ): Promise<boolean> => {
    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.service.list.error.no_host'));
        return false;
      }

      // 使用当前选中的分类
      if (!params.value.category || params.value.category.trim() === '') {
        Message.error(t('app.service.list.error.invalid_category'));
        return false;
      }

      await serviceActivateApi({
        type,
        category: params.value.category,
        name: record.name,
        action,
      });

      const actionText =
        action === SERVICE_ACTION.Activate
          ? t('app.service.list.operation.activate')
          : t('app.service.list.operation.deactivate');

      Message.success(
        t('app.service.list.success.action', { action: actionText })
      );
      return true;
    } catch (error) {
      console.error(`Failed to ${action} service:`, error);
      Message.error(t('app.service.list.error.action'));
      return false;
    }
  };

  // 服务操作（启动/停止/重启等）
  const operateService = async (
    record: ServiceEntity,
    operation: SERVICE_OPERATION
  ): Promise<string | null> => {
    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.service.list.error.no_host'));
        return null;
      }

      // 使用当前选中的分类
      if (!params.value.category || params.value.category.trim() === '') {
        Message.error(t('app.service.list.error.invalid_category'));
        return null;
      }

      const response = await serviceOperateApi({
        type,
        category: params.value.category,
        name: record.name,
        operation,
      });

      const operationText = t(
        `app.service.list.operation.${operation.toLowerCase()}`
      );
      Message.success(
        t('app.service.list.success.operation', { operation: operationText })
      );
      return response.result;
    } catch (error) {
      console.error(`Failed to ${operation} service:`, error);
      Message.error(t('app.service.list.error.operation'));
      return null;
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
    fetchServiceList,
    deleteService,
    toggleServiceStatus,
    operateService,
  };
}
