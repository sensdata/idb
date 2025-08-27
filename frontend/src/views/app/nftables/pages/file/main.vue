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
          <a-button :loading="loading" @click="handleRefresh">
            <template #icon>
              <icon-refresh />
            </template>
            {{ $t('app.nftables.button.refresh') }}
          </a-button>
          <a-button
            type="primary"
            :loading="saving"
            :disabled="!hasChanges || !content.trim()"
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
      <div class="editor-wrapper">
        <code-editor
          v-model="content"
          :autofocus="true"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="editorExtensions"
          :read-only="saving"
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

      <!-- 编辑器底部信息栏 -->
      <div class="editor-footer">
        <div class="editor-info">
          <a-space size="small">
            <a-tag :color="'rgb(var(--primary-6))'" size="small">
              <template #icon>
                <icon-file />
              </template>
              /etc/nftables.conf
            </a-tag>

            <a-tag
              v-if="content"
              :color="'rgb(var(--color-text-4))'"
              size="small"
            >
              {{
                $t('app.nftables.config.editor.lineCount', { count: lineCount })
              }}
            </a-tag>

            <a-tag
              v-if="hasChanges"
              :color="'rgb(var(--warning-6))'"
              size="small"
            >
              <template #icon>
                <icon-edit />
              </template>
              {{ $t('app.nftables.config.editor.modified') }}
            </a-tag>
          </a-space>
        </div>

        <div class="editor-tips">
          <a-typography-text type="secondary" class="tip-text">
            {{ $t('app.nftables.config.editor.tips') }}
          </a-typography-text>
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

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError, logInfo } = useLogger('NftablesFileEditor');

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
  const content = ref<string>('');
  const originalContent = ref<string>('');
  const editorView = ref<EditorView | null>(null);

  // 计算属性
  const hasChanges = computed(() => content.value !== originalContent.value);

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
      if (!saving.value && hasChanges.value && content.value.trim()) {
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
    justify-content: space-between;
    align-items: flex-start;
    padding: 16px 24px;
    background: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2);
    flex-shrink: 0;

    .header-left {
      .page-title {
        margin: 0 0 4px 0;
        font-size: 20px;
        font-weight: 600;
        color: var(--color-text-1);
      }

      .page-subtitle {
        color: var(--color-text-3);
        font-size: 14px;
        line-height: 1.5;
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
    align-items: center;
    justify-content: center;
    height: 400px;
    gap: 16px;

    .loading-text {
      color: var(--color-text-2);
      font-size: 14px;
    }
  }

  .error-container {
    padding: 40px 24px;
  }

  .editor-container {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    padding: 16px 24px;

    .editor-wrapper {
      position: relative;
      flex: 1;
      overflow: hidden;
      border-radius: 8px;

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
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        background: var(--color-mask-bg, rgba(255, 255, 255, 0.8));
        backdrop-filter: blur(2px);
        z-index: 10;
        gap: 12px;

        .saving-text {
          color: var(--color-text-2);
          font-size: 14px;
        }
      }
    }

    .editor-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      margin-top: 16px;
      background: var(--color-bg-2);
      border-radius: 6px;
      flex-shrink: 0;

      .editor-info {
        display: flex;
        align-items: center;
      }

      .editor-tips {
        display: flex;
        align-items: center;
      }

      .tip-text {
        font-size: 12px;
      }
    }
  }

  // 响应式设计
  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;

      .header-right {
        align-self: flex-end;
      }
    }

    .editor-container .editor-footer {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .editor-tips {
        align-self: flex-start;
      }
    }
  }

  @media (max-width: 480px) {
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
