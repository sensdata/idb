<template>
  <div class="source-editor-root">
    <div v-if="changedDirectives.length" class="change-summary">
      <a-alert type="info" :show-icon="true">
        {{
          $t('app.ssh.source.changeSummary', {
            count: changedDirectives.length,
            keys: changedDirectives.slice(0, 6).join(', '),
          })
        }}
      </a-alert>
    </div>
    <div v-if="riskMessages.length" class="risk-summary">
      <a-alert type="warning" :show-icon="true">
        <template #title>{{ $t('app.ssh.risk.title') }}</template>
        <div
          v-for="(message, index) in riskMessages"
          :key="`risk-${index}`"
          class="risk-item"
        >
          {{ message }}
        </div>
      </a-alert>
    </div>

    <div class="source-editor-wrapper">
      <div class="source-toolbar">
        <a-button
          type="primary"
          :disabled="!hasChanges || loading"
          :loading="loading"
          @click="handleSave"
        >
          {{ $t('app.ssh.source.save') }}
        </a-button>
        <a-button class="ml-2" :disabled="loading" @click="handleReset">
          {{ $t('app.ssh.source.reset') }}
        </a-button>
      </div>
      <div class="editor-container">
        <code-editor
          v-model="localConfig"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="extensions"
          :read-only="loading"
          :file="sshConfigFile"
          class="source-editor"
        />
      </div>
      <div class="source-info">
        {{ $t('app.ssh.source.info') }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { StreamLanguage } from '@codemirror/language';
  import { shell } from '@codemirror/legacy-modes/mode/shell';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useConfirm } from '@/composables/confirm';
  import CodeEditor from '@/components/code-editor/index.vue';

  const { t } = useI18n();
  const { confirm } = useConfirm();

  const props = defineProps({
    config: {
      type: String,
      required: true,
    },
    originalConfig: {
      type: String,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  });

  const emits = defineEmits(['update:config', 'save', 'reset', 'update-form']);

  const localConfig = ref(props.config);

  // 创建 SSH 配置文件对象
  const sshConfigFile = computed(() => ({
    name: 'sshd_config',
    path: '/etc/ssh/sshd_config',
  }));

  const extensions = [StreamLanguage.define(shell)];

  const trackedKeys = [
    'Port',
    'ListenAddress',
    'PermitRootLogin',
    'PasswordAuthentication',
    'PubkeyAuthentication',
    'UseDNS',
  ];

  const parseDirectives = (content: string): Record<string, string> => {
    const result: Record<string, string> = {};
    const lines = content.split('\n');

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith('#')) continue;
      const [key, ...rest] = trimmed.split(/\s+/);
      if (!trackedKeys.includes(key)) continue;
      result[key] = rest.join(' ').trim();
    }
    return result;
  };

  const originalDirectives = computed(() =>
    parseDirectives(props.originalConfig)
  );
  const currentDirectives = computed(() => parseDirectives(localConfig.value));

  const changedDirectives = computed(() => {
    const keys = new Set([
      ...Object.keys(originalDirectives.value),
      ...Object.keys(currentDirectives.value),
    ]);
    return [...keys].filter(
      (key) => originalDirectives.value[key] !== currentDirectives.value[key]
    );
  });

  const riskMessages = computed<string[]>(() => {
    const risks: string[] = [];
    const passwordAuth =
      (currentDirectives.value.PasswordAuthentication || '').toLowerCase() ===
      'yes';
    const keyAuth =
      (currentDirectives.value.PubkeyAuthentication || '').toLowerCase() ===
      'yes';
    const rootLogin = (
      currentDirectives.value.PermitRootLogin || ''
    ).toLowerCase();
    const port = Number(currentDirectives.value.Port || '');

    if (!passwordAuth && !keyAuth) {
      risks.push(t('app.ssh.risk.lockoutWarning'));
    }
    if (
      passwordAuth &&
      (rootLogin === 'yes' || rootLogin === 'without-password')
    ) {
      risks.push(t('app.ssh.risk.rootPasswordEnabled'));
    }
    if (
      currentDirectives.value.Port &&
      (!Number.isInteger(port) || port < 1 || port > 65535)
    ) {
      risks.push(t('app.ssh.risk.invalidPort'));
    }
    return risks;
  });

  // 计算是否有未保存的更改
  const hasChanges = computed(() => {
    return localConfig.value !== props.originalConfig;
  });

  // 监听 props.config 变化，更新 localConfig
  watch(
    () => props.config,
    (val) => {
      localConfig.value = val;
    }
  );

  // 监听 localConfig 变化，向父组件同步
  watch(localConfig, (val) => {
    emits('update:config', val);
  });

  // 处理保存按钮点击
  const handleSave = async () => {
    if (!localConfig.value.trim()) {
      Message.warning(t('app.ssh.source.emptyConfig'));
      return;
    }

    if (!hasChanges.value) {
      Message.info(t('app.ssh.source.noChanges'));
      return;
    }

    if (riskMessages.value.length) {
      const ok = await confirm({
        title: t('app.ssh.risk.confirmTitle'),
        content: riskMessages.value.join('\n'),
      });
      if (!ok) {
        return;
      }
    }

    emits('save');
  };

  // 处理重置按钮点击
  const handleReset = () => {
    localConfig.value = props.originalConfig;
    emits('reset');
  };

  // 检查是否有未保存的更改的公共方法
  const checkUnsavedChanges = () => {
    return hasChanges.value;
  };

  defineExpose({
    checkUnsavedChanges,
  });
</script>

<style scoped lang="less">
  .source-editor-root {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .source-editor-wrapper {
    display: flex;
    flex-direction: column;
    height: 500px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .source-toolbar {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-border-2);
    .ml-2 {
      margin-left: 8px;
    }
  }

  .editor-container {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .change-summary,
  .risk-summary {
    padding: 0;
  }

  .risk-item {
    margin-top: 2px;
  }

  .source-editor {
    position: absolute;
    inset: 0 0 0 0;
    width: 100%;
    height: 100%;
    font-family: Monaco, Menlo, 'Ubuntu Mono', Consolas, source-code-pro,
      monospace;
  }

  .source-editor :deep(.cm-editor) {
    height: 100%;
  }

  .source-editor :deep(.cm-scroller) {
    padding-bottom: 24px;
    overflow: auto;
  }

  .source-info {
    flex-shrink: 0;
    padding: 8px 12px;
    font-size: 12px;
    color: var(--color-text-3);
    border-top: 1px solid var(--color-border-2);
  }
</style>
