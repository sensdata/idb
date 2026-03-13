import { useI18n } from 'vue-i18n';
import { LOGROTATE_TYPE } from '@/config/enum';
import type { LogrotateEntity } from '@/entity/Logrotate';

export function useLogrotateActions() {
  const { t } = useI18n();

  // 获取状态显示信息
  const getStatusInfo = (record: LogrotateEntity) => {
    if (record.type === LOGROTATE_TYPE.System) {
      switch (record.sourceStatus) {
        case 'package_pristine':
          return {
            color: 'green',
            text: t('app.logrotate.list.source_status.package_pristine'),
          };
        case 'package_modified':
          return {
            color: 'orange',
            text: t('app.logrotate.list.source_status.package_modified'),
          };
        case 'package_owned':
          return {
            color: 'arcoblue',
            text: t('app.logrotate.list.source_status.package_owned'),
          };
        case 'local':
          return {
            color: 'gray',
            text: t('app.logrotate.list.source_status.local'),
          };
        default:
          return {
            color: 'gray',
            text: t('app.logrotate.list.source_status.unknown'),
          };
      }
    }

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
