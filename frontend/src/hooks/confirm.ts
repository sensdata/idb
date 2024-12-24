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

  function confirm(options: ConfirmOptions) {
    return new Promise((resolve) => {
      Modal.confirm({
        title: options.title || t('common.confirm.title'),
        content: options.content,
        okText: options.okText || t('common.confirm.okText'),
        cancelText: options.cancelText || t('common.confirm.cancelText'),
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
