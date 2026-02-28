<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.privateKeyInfo')"
    :width="820"
    :footer="false"
    class="private-key-drawer"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="privateKey" class="private-key-detail">
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="1"
          bordered
        >
          <a-descriptions-item :label="$t('app.certificate.alias')">
            {{ privateKey.alias }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.keyAlgorithm')">
            {{ privateKey.key_algorithm }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.keySize')">
            {{ privateKey.key_size }} bits
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.privateKeyPath')">
            <div class="path-cell">
              <a-typography-text code class="path-value">
                {{ privateKeyPath || '-' }}
              </a-typography-text>
              <div v-if="privateKeyPath" class="path-actions">
                <a-button type="text" size="small" @click="copyPrivateKeyPath">
                  <template #icon>
                    <icon-copy />
                  </template>
                  {{ $t('app.certificate.copyPrivateKeyPath') }}
                </a-button>
              </div>
            </div>
          </a-descriptions-item>
        </a-descriptions>

        <div class="section-block">
          <h4 class="section-title">{{
            $t('app.certificate.privateKeyContent')
          }}</h4>
          <a-textarea
            :model-value="formattedPem"
            :auto-size="{ minRows: 10, maxRows: 10 }"
            readonly
            class="key-content"
          />
        </div>

        <div class="action-bar">
          <a-button @click="copyPrivateKey">
            <template #icon>
              <icon-copy />
            </template>
            {{ $t('app.certificate.copyPrivateKey') }}
          </a-button>
          <a-button @click="drawerVisible = false">
            {{ $t('common.close') }}
          </a-button>
        </div>
      </div>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { IconCopy } from '@arco-design/web-vue/es/icon';
  import { useClipboard } from '@/composables/use-clipboard';
  import type { PrivateKeyInfo } from '@/api/certificate';

  interface Props {
    visible: boolean;
    privateKey?: PrivateKeyInfo | null;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  const { t } = useI18n();
  const { copyText } = useClipboard();

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  const privateKeyPath = computed(() => {
    const data = props.privateKey as
      | (PrivateKeyInfo & {
          path?: string;
          source?: string;
          key_path?: string;
          keyPath?: string;
        })
      | null
      | undefined;
    return data?.path || data?.source || data?.key_path || data?.keyPath || '';
  });

  const formattedPem = computed(() => {
    const pem = props.privateKey?.pem || '';
    return pem.includes('\\n') ? pem.replace(/\\n/g, '\n') : pem;
  });

  const copyPrivateKey = async () => {
    if (formattedPem.value) {
      try {
        await copyText(formattedPem.value);
        Message.success(t('app.certificate.copySuccess'));
      } catch (error) {
        Message.error(t('app.certificate.copyError'));
      }
    }
  };

  const copyPrivateKeyPath = async () => {
    if (!privateKeyPath.value) {
      Message.warning(t('app.certificate.pathUnavailable'));
      return;
    }
    try {
      await copyText(privateKeyPath.value);
      Message.success(t('app.certificate.copySuccess'));
    } catch (error) {
      Message.error(t('app.certificate.copyError'));
    }
  };
</script>

<style scoped>
  .private-key-detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-height: calc(100vh - 9rem);
    padding-bottom: 0.25rem;
    overflow-y: auto;
  }

  .path-value {
    max-width: 100%;
    overflow-wrap: anywhere;
  }

  .path-cell {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    align-items: flex-start;
  }

  .path-cell :deep(.arco-typography) {
    flex: 1;
    min-width: 0;
  }

  .path-actions {
    display: flex;
    gap: 0.375rem;
    align-items: center;
  }

  .section-block {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .section-title {
    margin: 0;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .key-content {
    font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
    font-size: 0.8125rem;
    line-height: 1.35;
  }

  .action-bar {
    position: sticky;
    bottom: 0;
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    padding-top: 0.75rem;
    margin-top: 0.25rem;
    background: linear-gradient(
      to bottom,
      rgb(255 255 255 / 10%),
      var(--color-bg-1) 40%
    );
  }

  :deep(.arco-descriptions-title) {
    margin-bottom: 0.75rem;
    font-size: 1rem;
    font-weight: 600;
  }

  :deep(.private-key-drawer .arco-drawer-body) {
    padding-bottom: 0.75rem;
  }
</style>
