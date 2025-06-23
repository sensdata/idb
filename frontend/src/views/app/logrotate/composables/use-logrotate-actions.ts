import { useI18n } from 'vue-i18n';
import type { LogrotateEntity } from '@/entity/Logrotate';

export function useLogrotateActions() {
  const { t } = useI18n();

  // 获取状态显示信息
  const getStatusInfo = (record: LogrotateEntity) => {
    return {
      color: record.linked ? 'green' : 'gray',
      text: record.linked
        ? t('app.logrotate.list.status.active')
        : t('app.logrotate.list.status.inactive'),
    };
  };

  // 获取激活/停用按钮信息
  const getActionButtonInfo = (record: LogrotateEntity) => {
    const isActive = record.linked;
    return {
      action: isActive ? ('deactivate' as const) : ('activate' as const),
      text: isActive
        ? t('app.logrotate.list.operation.deactivate')
        : t('app.logrotate.list.operation.activate'),
    };
  };

  return {
    getStatusInfo,
    getActionButtonInfo,
  };
}
