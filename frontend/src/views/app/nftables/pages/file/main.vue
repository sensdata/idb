<template>
  <div class="nftables-file-editor">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">{{ $t('app.nftables.config.file.title') }}</h1>
        <div class="page-subtitle">
          {{ $t('app.nftables.config.file.desc') }}
        </div>
      </div>
      <div class="header-right">
        <a-space>
          <a-button
            :disabled="loading || saving || hasError"
            @click="handleToggleEditMode"
          >
            <template #icon>
              <icon-edit v-if="!isEditMode" />
              <icon-lock v-else />
            </template>
            {{
              isEditMode
                ? $t('app.nftables.config.editor.exitEdit')
                : $t('app.nftables.config.editor.enableEdit')
            }}
          </a-button>
          <a-button :loading="loading" @click="handleRefresh">
            <template #icon>
              <icon-refresh />
            </template>
            {{ $t('app.nftables.button.refresh') }}
          </a-button>
          <a-button
            type="primary"
            :loading="saving"
            :disabled="editorReadOnly || !hasChanges || !content.trim()"
            @click="handleSave"
          >
            <template #icon>
              <icon-save />
            </template>
            {{ $t('app.nftables.button.save') }}
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !content" class="loading-container">
      <a-spin :size="32" />
      <span class="loading-text">{{ $t('common.loading') }}</span>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="hasError" class="error-container">
      <a-result status="error" :title="$t('common.error.title')">
        <template #subtitle>
          {{ errorMessage }}
        </template>
        <template #extra>
          <a-button type="primary" @click="handleRefresh">
            {{ $t('common.button.retry') }}
          </a-button>
        </template>
      </a-result>
    </div>

    <!-- 编辑器 -->
    <div v-else class="editor-container">
      <div class="editor-topbar">
        <div class="editor-info">
          <div class="meta-chip file-chip">
            <icon-file />
            <span>/etc/nftables.conf</span>
          </div>
          <div v-if="content" class="meta-chip">
            {{
              $t('app.nftables.config.editor.lineCount', { count: lineCount })
            }}
          </div>
          <div
            class="meta-chip mode-chip"
            :class="isEditMode ? 'editing' : 'readonly'"
          >
            {{
              $t(
                isEditMode
                  ? 'app.nftables.config.editor.modeEditing'
                  : 'app.nftables.config.editor.modeReadOnly'
              )
            }}
          </div>
          <div v-if="hasChanges" class="meta-chip modified-chip">
            <icon-edit />
            <span>{{ $t('app.nftables.config.editor.modified') }}</span>
          </div>
        </div>
        <div class="editor-tips">
          <a-typography-text type="secondary" class="tip-text">
            {{
              isEditMode
                ? $t('app.nftables.config.editor.tips')
                : $t('app.nftables.config.editor.doubleClickToEdit')
            }}
          </a-typography-text>
        </div>
      </div>

      <div
        class="editor-wrapper"
        :class="{ readonly: !isEditMode }"
        @dblclick="handleEnableEditMode"
      >
        <code-editor
          v-model="content"
          :autofocus="true"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="editorExtensions"
          :read-only="editorReadOnly"
          :file="nftablesFile"
          class="nftables-editor"
          @editor-ready="handleEditorReady"
        />

        <!-- 保存中的遮罩层 -->
        <div v-if="saving" class="saving-overlay">
          <a-spin :size="24" />
          <span class="saving-text">{{ $t('common.saving') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
  import { EditorView } from '@codemirror/view';
  import { Message, Modal } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import {
    IconRefresh,
    IconSave,
    IconFile,
    IconEdit,
    IconLock,
  } from '@arco-design/web-vue/es/icon';
  import { useLogger } from '@/composables/use-logger';
  import useCurrentHost from '@/composables/current-host';
  import useEditorConfig from '@/components/code-editor/composables/use-editor-config';
  import CodeEditor from '@/components/code-editor/index.vue';
  import {
    getNftablesRawConfigApi,
    updateNftablesRawConfigApi,
    type NftablesRawConfig,
  } from '@/api/nftables';
  import { useNftablesConfig } from '../../composables/use-nftables-config';

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError, logInfo } = useLogger('NftablesFileEditor');

  // 获取配置应用后的刷新方法（关闭自动请求避免重复）
  const { handleConfigApplied } = useNftablesConfig({ autoFetch: false });

  // 当前主机
  const { currentHostId } = useCurrentHost();

  // 编辑器配置
  const { getNftablesExtensions } = useEditorConfig(ref(null));

  // 创建 nftables 文件对象
  const nftablesFile = computed(() => ({
    name: 'nftables.conf',
    path: '/etc/nftables.conf',
  }));

  // 响应式状态
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const hasError = ref<boolean>(false);
  const errorMessage = ref<string>('');
  const isEditMode = ref<boolean>(false);
  const content = ref<string>('');
  const originalContent = ref<string>('');
  const editorView = ref<EditorView | null>(null);

  // 计算属性
  const hasChanges = computed(() => content.value !== originalContent.value);
  const editorReadOnly = computed(() => saving.value || !isEditMode.value);

  const lineCount = computed(() => {
    if (!content.value) return 0;
    return content.value.split('\n').length;
  });

  // CodeMirror 扩展配置
  const editorExtensions = computed(() => [...getNftablesExtensions()]);

  // 获取 nftables 原始配置
  const fetchNftablesRawConfig = async (): Promise<void> => {
    if (!currentHostId.value) {
      logError('No host selected');
      hasError.value = true;
      errorMessage.value = t('common.error.noHostSelected');
      return;
    }

    try {
      loading.value = true;
      hasError.value = false;
      errorMessage.value = '';

      logInfo('Fetching nftables raw config for host:', currentHostId.value);

      const response = await getNftablesRawConfigApi();

      content.value = response.content || '';
      originalContent.value = response.content || '';
      isEditMode.value = false;

      logInfo('Successfully loaded nftables raw config');
    } catch (error) {
      logError('Failed to fetch nftables raw config:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.fetchConfigFailed');
      Message.error(t('app.nftables.message.loadFailed'));
    } finally {
      loading.value = false;
    }
  };

  // 保存 nftables 原始配置
  const saveNftablesRawConfig = async (): Promise<void> => {
    if (!currentHostId.value) {
      logError('No host selected');
      return;
    }

    if (!content.value.trim()) {
      Message.warning(t('app.nftables.config.editor.emptyContent'));
      return;
    }

    try {
      saving.value = true;
      hasError.value = false;

      logInfo('Saving nftables raw config for host:', currentHostId.value);

      const updateData: NftablesRawConfig = {
        content: content.value,
      };

      await updateNftablesRawConfigApi(updateData);

      // 更新原始内容，清除修改状态
      originalContent.value = content.value;

      Message.success(t('app.nftables.message.configSaved'));
      logInfo('Successfully saved nftables raw config');

      // 配置应用成功后刷新当前配置列表
      await handleConfigApplied();
    } catch (error) {
      logError('Failed to save nftables raw config:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.saveConfigFailed');
      Message.error(t('app.nftables.message.saveFailed'));
    } finally {
      saving.value = false;
    }
  };

  // 刷新配置
  const handleRefresh = async (): Promise<void> => {
    if (hasChanges.value) {
      Modal.confirm({
        title: t('app.nftables.config.editor.confirmRefresh'),
        content: t('app.nftables.config.editor.confirmRefreshContent'),
        onOk: async () => {
          await fetchNftablesRawConfig();
        },
      });
    } else {
      await fetchNftablesRawConfig();
    }
  };

  // 保存配置
  const handleSave = async (): Promise<void> => {
    await saveNftablesRawConfig();
  };

  // 编辑器就绪事件
  const handleEditorReady = (payload: any): void => {
    editorView.value = payload.view;
    logInfo('Editor ready');
  };

  // 切换编辑模式
  const handleToggleEditMode = (): void => {
    if (saving.value || loading.value || hasError.value) return;

    if (!isEditMode.value) {
      isEditMode.value = true;
      editorView.value?.focus();
      return;
    }

    if (hasChanges.value) {
      Modal.confirm({
        title: t('app.nftables.config.editor.confirmExitEdit'),
        content: t('app.nftables.config.editor.confirmExitEditContent'),
        onOk: () => {
          content.value = originalContent.value;
          isEditMode.value = false;
        },
      });
      return;
    }

    isEditMode.value = false;
  };

  // 双击编辑区进入编辑模式
  const handleEnableEditMode = (): void => {
    if (saving.value || loading.value || hasError.value || isEditMode.value) {
      return;
    }
    isEditMode.value = true;
    editorView.value?.focus();
  };

  // 页面离开前的确认
  const handleBeforeUnload = (event: BeforeUnloadEvent): void => {
    if (hasChanges.value) {
      event.preventDefault();
      event.returnValue = t('app.nftables.config.editor.unsavedChanges');
    }
  };

  // 键盘快捷键
  const handleKeyDown = (event: KeyboardEvent): void => {
    // Ctrl+S 保存
    if ((event.ctrlKey || event.metaKey) && event.key === 's') {
      event.preventDefault();
      if (!isEditMode.value) {
        Message.warning(t('app.nftables.config.editor.readOnlyNoSave'));
        return;
      }
      if (
        isEditMode.value &&
        !saving.value &&
        hasChanges.value &&
        content.value.trim()
      ) {
        handleSave();
      }
    }
  };

  // 组件挂载时的处理
  onMounted(async () => {
    await fetchNftablesRawConfig();

    // 注册页面离开前的确认事件
    window.addEventListener('beforeunload', handleBeforeUnload);

    // 注册键盘快捷键
    document.addEventListener('keydown', handleKeyDown);
  });

  // 组件卸载时的清理
  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload);
    document.removeEventListener('keydown', handleKeyDown);
  });
</script>

<style scoped lang="less">
  .nftables-file-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .page-header {
    display: flex;
    flex-shrink: 0;
    align-items: flex-start;
    justify-content: space-between;
    padding: 16px 24px;
    background: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2);
    .header-left {
      .page-title {
        margin: 0;
        font-size: 1.286rem;
        font-weight: 500;
        color: var(--color-text-1);
      }
      .page-subtitle {
        font-size: 14px;
        line-height: 1.5;
        color: var(--color-text-3);
      }
    }
    .header-right {
      display: flex;
      align-items: center;
    }
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    align-items: center;
    justify-content: center;
    height: 400px;
    .loading-text {
      font-size: 14px;
      color: var(--color-text-2);
    }
  }

  .error-container {
    padding: 40px 24px;
  }

  .editor-container {
    display: flex;
    flex: 1;
    flex-direction: column;
    padding: 16px 24px;
    overflow: hidden;
    .editor-topbar {
      display: flex;
      flex-shrink: 0;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      padding: 10px 12px;
      margin-bottom: 12px;
      background: var(--color-bg-2);
      border: 1px solid var(--color-border-2);
      border-radius: 6px;
      .editor-info {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        align-items: center;
      }
      .meta-chip {
        display: inline-flex;
        gap: 6px;
        align-items: center;
        padding: 3px 10px;
        font-size: 12px;
        color: var(--color-text-2);
        background: var(--color-fill-2);
        border: 1px solid var(--color-border-2);
        border-radius: 999px;
      }
      .file-chip {
        color: rgb(var(--primary-6));
        background: rgb(var(--primary-1));
        border-color: rgb(var(--primary-3));
      }
      .mode-chip.readonly {
        color: rgb(var(--warning-7));
        background: rgb(var(--warning-1));
        border-color: rgb(var(--warning-3));
      }
      .mode-chip.editing {
        color: rgb(var(--success-7));
        background: rgb(var(--success-1));
        border-color: rgb(var(--success-3));
      }
      .modified-chip {
        color: rgb(var(--warning-7));
        background: rgb(var(--warning-1));
        border-color: rgb(var(--warning-3));
      }
      .editor-tips {
        display: flex;
        flex: 1;
        align-items: center;
        justify-content: flex-end;
      }
      .tip-text {
        font-size: 12px;
        white-space: nowrap;
      }
    }
    .editor-wrapper {
      position: relative;
      flex: 1;
      overflow: hidden;
      border-radius: 8px;
      &.readonly {
        :deep(.cm-content) {
          cursor: default;
        }
      }
      .nftables-editor {
        width: 100%;
        height: 100%;
        min-height: 500px;
        :deep(.cm-editor) {
          width: 100%;
          height: 100%;
        }
        :deep(.cm-scroller) {
          width: 100%;
          height: 100%;
          overflow: auto;
        }
        :deep(.cm-content) {
          min-height: 500px;
        }
      }
      .saving-overlay {
        position: absolute;
        inset: 0 0 0 0;
        z-index: 10;
        display: flex;
        flex-direction: column;
        gap: 12px;
        align-items: center;
        justify-content: center;
        background: var(--color-mask-bg, rgb(255 255 255 / 80%));
        backdrop-filter: blur(2px);
        .saving-text {
          font-size: 14px;
          color: var(--color-text-2);
        }
      }
    }
  }

  // 响应式设计
  @media (width <= 768px) {
    .page-header {
      flex-direction: column;
      gap: 16px;
      align-items: flex-start;
      .header-right {
        align-self: flex-end;
      }
    }
    .editor-container .editor-topbar {
      flex-direction: column;
      gap: 12px;
      align-items: flex-start;
      .editor-tips {
        align-self: flex-start;
      }
    }
  }

  @media (width <= 480px) {
    .page-header {
      padding: 12px 16px;
      .header-left .page-title {
        font-size: 18px;
      }
    }
    .editor-container {
      padding: 12px 16px;
    }
  }
</style>
