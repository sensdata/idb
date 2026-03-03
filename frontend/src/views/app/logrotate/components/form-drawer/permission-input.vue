<template>
  <div class="permission-input">
    <div class="permission-editor">
      <a-row :gutter="12">
        <a-col :span="8">
          <div class="editor-label">{{
            $t('app.logrotate.permission.mode')
          }}</div>
          <a-input
            :model-value="modeInput"
            :placeholder="$t('app.logrotate.permission.mode_placeholder')"
            :disabled="disabled"
            @update:model-value="handleModeInput"
            @blur="handleModeBlur"
          />
        </a-col>
        <a-col :span="8">
          <div class="editor-label">{{
            $t('app.logrotate.permission.user')
          }}</div>
          <a-input
            :model-value="userInput"
            :placeholder="$t('app.logrotate.permission.user_placeholder')"
            :disabled="disabled"
            @update:model-value="handleUserInput"
          />
        </a-col>
        <a-col :span="8">
          <div class="editor-label">
            {{ $t('app.logrotate.permission.group_name') }}
          </div>
          <a-input
            :model-value="groupInput"
            :placeholder="$t('app.logrotate.permission.group_placeholder')"
            :disabled="disabled"
            @update:model-value="handleGroupInput"
          />
        </a-col>
      </a-row>

      <div class="advanced-toggle-row">
        <a-button
          type="text"
          size="small"
          :disabled="disabled"
          @click="showAdvanced = !showAdvanced"
        >
          {{
            showAdvanced
              ? $t('app.logrotate.permission.advanced_hide')
              : $t('app.logrotate.permission.advanced_show')
          }}
        </a-button>
      </div>

      <a-row v-if="showAdvanced" class="permission-checkboxes" :gutter="12">
        <a-col :span="8">
          <div class="editor-label">{{
            $t('app.logrotate.permission.owner')
          }}</div>
          <a-checkbox-group
            v-model="accessState.owner"
            :options="accessOptions"
            :disabled="disabled"
            @change="handleAccessChange"
          />
        </a-col>
        <a-col :span="8">
          <div class="editor-label">{{
            $t('app.logrotate.permission.group')
          }}</div>
          <a-checkbox-group
            v-model="accessState.group"
            :options="accessOptions"
            :disabled="disabled"
            @change="handleAccessChange"
          />
        </a-col>
        <a-col :span="8">
          <div class="editor-label">{{
            $t('app.logrotate.permission.other')
          }}</div>
          <a-checkbox-group
            v-model="accessState.other"
            :options="accessOptions"
            :disabled="disabled"
            @change="handleAccessChange"
          />
        </a-col>
      </a-row>

      <div class="permission-preview-compact">
        <div class="preview-content">
          <div class="preview-label">
            {{ $t('app.logrotate.permission.preview') }}:
          </div>
          <div class="preview-value-with-comment">
            <code class="preview-value">{{ previewCreate }}</code>
            <span class="preview-comment">// {{ previewDescription }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { usePermission } from './composables/use-permission';
  import type { PermissionAccess } from './types';

  interface Props {
    modelValue?: string;
    disabled?: boolean;
  }

  interface Emits {
    (e: 'update:modelValue', value: string): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: 'create 0644 root root',
    disabled: false,
  });

  const emit = defineEmits<Emits>();
  const { t } = useI18n();

  const {
    permission,
    access,
    updateFromString,
    updateFromAccess,
    toString,
    parseCreateString,
  } = usePermission(props.modelValue);

  const showAdvanced = ref(false);
  const disabled = computed(() => props.disabled);
  const modeInput = ref('0644');
  const userInput = ref('root');
  const groupInput = ref('root');
  const accessState = ref<PermissionAccess>({
    owner: [],
    group: [],
    other: [],
  });

  const accessOptions = computed(() => [
    { label: t('app.logrotate.permission.read'), value: '4' },
    { label: t('app.logrotate.permission.write'), value: '2' },
    { label: t('app.logrotate.permission.execute'), value: '1' },
  ]);

  const previewCreate = computed(
    () => `create ${modeInput.value} ${userInput.value} ${groupInput.value}`
  );

  const previewDescription = computed(() => {
    if (!/^0?[0-7]{3,4}$/.test(modeInput.value)) {
      return t('app.logrotate.permission.invalid_mode');
    }
    return permission.value.description;
  });

  const calculateMode = (values: string[]): number =>
    values.reduce((sum, current) => sum + Number(current), 0);

  const syncByFields = () => {
    const normalizedMode = modeInput.value.padStart(4, '0');
    const normalizedUser = userInput.value.trim() || 'root';
    const normalizedGroup = groupInput.value.trim() || 'root';
    const createValue = `create ${normalizedMode} ${normalizedUser} ${normalizedGroup}`;

    updateFromString(createValue);
    emit('update:modelValue', toString());
  };

  const syncFromModelValue = (value?: string) => {
    const parsed = parseCreateString(value || '');
    modeInput.value = parsed.mode;
    userInput.value = parsed.user;
    groupInput.value = parsed.group;
    accessState.value = {
      owner: [...parsed.access.owner],
      group: [...parsed.access.group],
      other: [...parsed.access.other],
    };
    updateFromString(`create ${parsed.mode} ${parsed.user} ${parsed.group}`);
  };

  const handleModeInput = (value: string) => {
    modeInput.value = value.replace(/[^0-7]/g, '').slice(0, 4);
    if (/^[0-7]{3,4}$/.test(modeInput.value)) {
      syncByFields();
    }
  };

  const handleModeBlur = () => {
    if (/^[0-7]{3}$/.test(modeInput.value)) {
      modeInput.value = `0${modeInput.value}`;
      syncByFields();
    }
  };

  const handleUserInput = (value: string) => {
    userInput.value = value;
    syncByFields();
  };

  const handleGroupInput = (value: string) => {
    groupInput.value = value;
    syncByFields();
  };

  const handleAccessChange = () => {
    const owner = calculateMode(accessState.value.owner);
    const group = calculateMode(accessState.value.group);
    const other = calculateMode(accessState.value.other);
    modeInput.value = `0${owner}${group}${other}`;
    updateFromAccess(accessState.value);
    emit(
      'update:modelValue',
      toString({
        mode: modeInput.value,
        user: userInput.value.trim() || 'root',
        group: groupInput.value.trim() || 'root',
        description: '',
      })
    );
  };

  watch(
    access,
    (newAccess) => {
      accessState.value = {
        owner: [...newAccess.owner],
        group: [...newAccess.group],
        other: [...newAccess.other],
      };
    },
    { deep: true }
  );

  watch(
    () => props.modelValue,
    (newValue: string) => {
      syncFromModelValue(newValue);
    },
    { immediate: true }
  );
</script>

<style scoped>
  .permission-input {
    width: 100%;
  }

  .permission-editor {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 10px 12px;
    background-color: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .editor-label {
    margin-bottom: 6px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .advanced-toggle-row {
    display: flex;
    justify-content: flex-end;
  }

  .permission-checkboxes {
    padding: 8px;
    background: var(--color-bg-1);
    border: 1px dashed var(--color-border-2);
    border-radius: 6px;
  }

  :deep(.arco-checkbox-group) {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  :deep(.arco-checkbox) {
    margin-right: 0;
  }

  .permission-preview-compact {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    padding: 10px 12px;
    background-color: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .preview-content {
    flex: 1;
  }

  .preview-label {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .preview-value {
    display: inline;
    padding: 4px 8px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 13px;
    color: var(--color-text-1);
    background-color: var(--color-bg-3);
    border-radius: 3px;
  }

  .preview-value-with-comment {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
  }

  .preview-comment {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    font-style: italic;
    color: var(--color-text-3);
  }
</style>
