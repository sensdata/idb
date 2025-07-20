<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.privateKeyInfo')"
    :width="500"
    :footer="false"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="privateKey" class="private-key-detail">
        <!-- 基本信息 -->
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="2"
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
        </a-descriptions>

        <!-- 私钥内容 -->
        <div class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.privateKeyContent') }}</h4>
          <a-textarea
            :model-value="privateKey.pem"
            :rows="12"
            readonly
            class="font-mono text-sm"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="mt-6 flex justify-end gap-3">
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

  // Props 定义
  interface Props {
    visible: boolean;
    privateKey?: PrivateKeyInfo | null;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  const { t } = useI18n();
  const { copyText } = useClipboard();

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 复制私钥内容
  const copyPrivateKey = async () => {
    if (props.privateKey?.pem) {
      try {
        await copyText(props.privateKey.pem);
        Message.success(t('app.certificate.copySuccess'));
      } catch (error) {
        Message.error(t('app.certificate.copyError'));
      }
    }
  };
</script>

<style scoped>
  .private-key-detail {
    max-height: 70vh;
    overflow-y: auto;
  }

  :deep(.arco-descriptions-title) {
    margin-bottom: 1rem;
    font-size: 1.33rem;
    font-weight: 600;
  }

  :deep(.arco-textarea) {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
  }

  .mt-6 {
    margin-top: 2rem;
  }

  .mb-3 {
    margin-bottom: 1rem;
  }

  h4 {
    margin: 0;
    font-size: 1.33rem;
    font-weight: 600;
  }
</style>
