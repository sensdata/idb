<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.service.form.environment.editor.title')"
    :width="600"
    unmount-on-close
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <div class="environment-editor">
      <div class="editor-header">
        <span class="editor-description">
          {{ $t('app.service.form.environment.editor.description') }}
        </span>
        <a-button type="primary" size="small" @click="addEnvironment">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.service.form.environment.editor.add') }}
        </a-button>
      </div>

      <div class="environment-list">
        <div
          v-for="(env, index) in environments"
          :key="index"
          class="environment-item"
        >
          <div class="env-inputs">
            <a-input
              v-model="env.key"
              :placeholder="$t('app.service.form.environment.editor.key')"
              class="env-key"
            />
            <span class="env-separator">=</span>
            <a-input
              v-model="env.value"
              :placeholder="$t('app.service.form.environment.editor.value')"
              class="env-value"
            />
          </div>
          <a-button
            type="text"
            status="danger"
            size="small"
            @click="removeEnvironment(index)"
          >
            <template #icon>
              <icon-delete />
            </template>
          </a-button>
        </div>

        <div v-if="environments.length === 0" class="empty-state">
          <icon-folder-add />
          <span>{{ $t('app.service.form.environment.editor.empty') }}</span>
        </div>
      </div>

      <div class="editor-footer">
        <div class="preview-section">
          <div class="preview-title">
            {{ $t('app.service.form.environment.editor.preview') }}
          </div>
          <div class="preview-content">
            <code>{{ envPreview }}</code>
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';

  interface EnvironmentVariable {
    key: string;
    value: string;
  }

  const emit = defineEmits<{
    update: [value: string];
  }>();

  useI18n();

  const visible = ref(false);
  const environments = ref<EnvironmentVariable[]>([]);

  // 计算预览字符串
  const envPreview = computed(() => {
    return environments.value
      .filter((env) => env.key.trim() && env.value.trim())
      .map((env) => {
        const key = env.key.trim();
        const value = env.value.trim();

        // 如果键或值包含空格或特殊字符，需要加引号
        const needsQuotesForKey =
          key.includes(' ') || key.includes('"') || key.includes("'");
        const needsQuotesForValue =
          value.includes(' ') || value.includes('"') || value.includes("'");

        const quotedKey = needsQuotesForKey
          ? `"${key.replace(/"/g, '\\"')}"`
          : key;
        const quotedValue = needsQuotesForValue
          ? `"${value.replace(/"/g, '\\"')}"`
          : value;

        return `${quotedKey}=${quotedValue}`;
      })
      .join(' ');
  });

  // 添加环境变量
  const addEnvironment = () => {
    environments.value.push({ key: '', value: '' });
  };

  // 删除环境变量
  const removeEnvironment = (index: number) => {
    environments.value.splice(index, 1);
  };

  // 从字符串解析环境变量
  const parseEnvironmentString = (envString: string) => {
    if (!envString.trim()) {
      return [];
    }

    // 解析单个键值对
    const parseSinglePair = (pairStr: string): EnvironmentVariable | null => {
      const equalIndex = pairStr.indexOf('=');
      if (equalIndex === -1) return null;

      let key = pairStr.substring(0, equalIndex).trim();
      let value = pairStr.substring(equalIndex + 1).trim();

      // 移除键的引号
      if (
        (key.startsWith('"') && key.endsWith('"')) ||
        (key.startsWith("'") && key.endsWith("'"))
      ) {
        key = key.slice(1, -1);
      }

      // 移除值的引号
      if (
        (value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'"))
      ) {
        value = value.slice(1, -1);
      }

      return { key, value };
    };

    // 更智能的解析逻辑，正确处理引号
    const parseEnvString = (str: string): EnvironmentVariable[] => {
      const result: EnvironmentVariable[] = [];
      let current = '';
      let inQuotes = false;
      let quoteChar = '';

      for (let i = 0; i < str.length; i++) {
        const char = str[i];

        if ((char === '"' || char === "'") && !inQuotes) {
          // 开始引用
          inQuotes = true;
          quoteChar = char;
          // 不将引号添加到current中，因为我们想解析引号内的内容
        } else if (char === quoteChar && inQuotes) {
          // 结束引用
          inQuotes = false;
          quoteChar = '';
          // 不将引号添加到current中
        } else if (char === ' ' && !inQuotes) {
          // 空格分隔符（不在引号内）
          if (current.trim()) {
            const pair = parseSinglePair(current.trim());
            if (pair) result.push(pair);
            current = '';
          }
        } else {
          current += char;
        }
      }

      // 处理最后一个
      if (current.trim()) {
        const pair = parseSinglePair(current.trim());
        if (pair) result.push(pair);
      }

      return result;
    };

    return parseEnvString(envString);
  };

  // 显示弹窗
  const show = (currentValue = '') => {
    visible.value = true;
    environments.value = parseEnvironmentString(currentValue);

    // 如果没有环境变量，添加一个空的
    if (environments.value.length === 0) {
      addEnvironment();
    }
  };

  // 确认
  const handleOk = () => {
    emit('update', envPreview.value);
    visible.value = false;
  };

  // 取消
  const handleCancel = () => {
    visible.value = false;
  };

  defineExpose({
    show,
  });
</script>

<style scoped>
  .environment-editor {
    padding: 8px 0;
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    margin-bottom: 16px;
    border-bottom: 1px solid var(--color-border-2);
  }

  .editor-description {
    font-size: 14px;
    color: var(--color-text-2);
  }

  .environment-list {
    min-height: 200px;
    max-height: 300px;
    overflow-y: auto;
  }

  .environment-item {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 8px;
    margin-bottom: 12px;
    background-color: var(--color-fill-1);
    border-radius: 4px;
  }

  .env-inputs {
    display: flex;
    flex: 1;
    gap: 8px;
    align-items: center;
  }

  .env-key {
    flex: 1;
    min-width: 120px;
  }

  .env-separator {
    font-weight: bold;
    color: var(--color-text-2);
  }

  .env-value {
    flex: 2;
    min-width: 180px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    font-size: 14px;
    color: var(--color-text-3);
  }

  .empty-state .arco-icon {
    margin-bottom: 12px;
    font-size: 48px;
    opacity: 0.5;
  }

  .editor-footer {
    padding-top: 12px;
    margin-top: 16px;
    border-top: 1px solid var(--color-border-2);
  }

  .preview-section {
    margin-bottom: 8px;
  }

  .preview-title {
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .preview-content {
    padding: 8px;
    background-color: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .preview-content code {
    font-family: 'Courier New', monospace;
    font-size: 12px;
    color: var(--color-text-1);
    word-break: break-all;
  }
</style>
