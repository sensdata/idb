import { RenderFunction } from 'vue';
import { Modal } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';

export interface ConfirmOptions {
  title?: string;
  content: string | RenderFunction;
  okText?: string;
  cancelText?: string;
}

export function useConfirm() {
  const { t } = useI18n();

  function confirm(options: ConfirmOptions | string) {
    if (typeof options === 'string') {
      options = {
        content: options,
      };
    }
    return new Promise((resolve) => {
      Modal.confirm({
        title: (options as ConfirmOptions).title || t('common.confirm.title'),
        content: (options as ConfirmOptions).content,
        okText:
          (options as ConfirmOptions).okText || t('common.confirm.okText'),
        cancelText:
          (options as ConfirmOptions).cancelText ||
          t('common.confirm.cancelText'),
        onCancel: () => {
          resolve(false);
        },
        onOk: () => {
          resolve(true);
        },
      });
    });
  }

  return { confirm };
}
