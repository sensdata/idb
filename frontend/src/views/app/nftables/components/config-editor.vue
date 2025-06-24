<template>
  <div class="config-editor">
    <div class="editor-header">
      <div class="editor-title">
        <h3>{{ $t('app.nftables.config.file.title') }}</h3>
        <a-tag :color="isConfigExist ? 'green' : 'orange'" size="small">
          {{
            isConfigExist
              ? $t('app.nftables.config.file.exists')
              : $t('app.nftables.config.file.new')
          }}
        </a-tag>
      </div>

      <div class="editor-actions">
        <a-space>
          <a-select
            :model-value="configType"
            size="small"
            class="config-type-select"
            @change="handleConfigTypeChange"
          >
            <a-option value="local">{{
              $t('app.nftables.config.type.local')
            }}</a-option>
            <a-option value="global">{{
              $t('app.nftables.config.type.global')
            }}</a-option>
          </a-select>

          <a-button size="small" @click="handleRefresh">
            <template #icon>
              <icon-refresh />
            </template>
            {{ $t('app.nftables.button.refresh') }}
          </a-button>

          <a-button
            type="primary"
            size="small"
            :loading="saving"
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

    <div class="editor-content">
      <div class="editor-wrapper" :class="{ loading }">
        <codemirror
          :model-value="content"
          :placeholder="$t('app.nftables.config.editor.placeholder')"
          :autofocus="true"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="nftablesExtensions"
          class="config-codemirror"
          @update:model-value="handleContentChange"
        />
        <div v-if="loading" class="editor-loading">
          <a-spin :size="32" />
        </div>
      </div>
    </div>

    <div class="editor-footer">
      <div class="editor-info">
        <a-space size="small">
          <a-tag color="blue" size="small">
            <template #icon>
              <icon-file />
            </template>
            {{ configFilePath }}
          </a-tag>

          <a-tag v-if="content" color="gray" size="small">
            {{
              $t('app.nftables.config.editor.lineCount', { count: lineCount })
            }}
          </a-tag>
        </a-space>
      </div>

      <div class="editor-tips">
        <a-typography-text type="secondary" class="tip-text">
          {{ $t('app.nftables.config.editor.reloadTip') }}
        </a-typography-text>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { EditorView } from '@codemirror/view';
  import { basicSetup } from 'codemirror';
  import {
    IconRefresh,
    IconSave,
    IconFile,
  } from '@arco-design/web-vue/es/icon';
  import type { ConfigType } from '@/api/nftables';
  import useEditorConfig from '@/components/code-editor/composables/use-editor-config';

  interface Props {
    content: string;
    configType: ConfigType;
    isConfigExist: boolean;
    loading?: boolean;
    saving?: boolean;
  }

  interface Emits {
    (e: 'update:content', content: string): void;
    (e: 'update:configType', type: ConfigType): void;
    (e: 'refresh'): void;
    (e: 'save'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    saving: false,
  });

  const emit = defineEmits<Emits>();

  // 使用编辑器配置 hook
  const { getNftablesExtensions } = useEditorConfig(ref(null));

  // CodeMirror 扩展配置
  const nftablesExtensions = computed(() => [
    basicSetup,
    ...getNftablesExtensions(),
    EditorView.theme({
      '&': {
        fontSize: '14px',
        fontFamily:
          'Monaco, Menlo, Ubuntu Mono, Consolas, source-code-pro, monospace',
      },
      '.cm-content': {
        padding: '16px',
        minHeight: '400px',
        lineHeight: '1.5',
      },
      '.cm-focused': {
        outline: 'none',
      },
      '.cm-editor': {
        borderRadius: '8px',
        border: '1px solid var(--color-border-2)',
        backgroundColor: 'var(--color-bg-2)',
      },
      '.cm-editor.cm-focused': {
        borderColor: 'var(--color-primary-light-4)',
        boxShadow: '0 0 0 2px var(--color-primary-light-8)',
      },
      '.cm-scroller': {
        fontFamily:
          'Monaco, Menlo, Ubuntu Mono, Consolas, source-code-pro, monospace',
      },
    }),
    EditorView.lineWrapping,
  ]);

  // 计算属性
  const lineCount = computed(() => {
    return props.content ? props.content.split('\n').length : 0;
  });

  const configFilePath = computed(() => {
    return `/nftables/${props.configType}/ports/main.nft`;
  });

  // 事件处理
  const handleContentChange = (value: string) => {
    emit('update:content', value);
  };

  const handleConfigTypeChange = (
    type:
      | string
      | number
      | boolean
      | Record<string, any>
      | (string | number | boolean | Record<string, any>)[]
  ) => {
    emit('update:configType', type as ConfigType);
  };

  const handleRefresh = () => {
    emit('refresh');
  };

  const handleSave = () => {
    emit('save');
  };
</script>

<style scoped lang="less">
  .config-editor {
    .editor-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      padding: 16px;
      background: var(--color-bg-2);
      border-radius: 8px;
      border: 1px solid var(--color-border-2);

      .editor-title {
        display: flex;
        align-items: center;
        gap: 12px;

        h3 {
          font-size: 16px;
          font-weight: 500;
          color: var(--color-text-1);
          margin: 0;
        }
      }

      .editor-actions {
        display: flex;
        align-items: center;
      }
    }

    .config-type-select {
      width: 100px;
    }

    .editor-content {
      margin-bottom: 16px;

      .editor-wrapper {
        position: relative;
        min-height: 400px;

        &.loading {
          .config-codemirror {
            opacity: 0.6;
            pointer-events: none;
          }
        }

        .config-codemirror {
          width: 100%;
          min-height: 400px;
          transition: opacity 0.3s ease;

          :deep(.cm-editor) {
            width: 100%;
            min-height: 400px;
          }

          :deep(.cm-scroller) {
            min-height: 400px;
          }
        }

        .editor-loading {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          z-index: 10;
        }
      }
    }

    .editor-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      background: var(--color-bg-3);
      border-radius: 6px;

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
    .config-editor {
      .editor-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 12px;

        .editor-actions {
          align-self: flex-end;
        }
      }

      .editor-footer {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;

        .editor-tips {
          align-self: flex-start;
        }
      }
    }
  }

  @media (max-width: 480px) {
    .config-editor {
      .editor-header .editor-title {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
      }
    }
  }
</style>
