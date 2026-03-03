<template>
  <a-form
    ref="formRef"
    :model="formData"
    :rules="formRules"
    layout="vertical"
    hide-asterisk
  >
    <section class="summary-section">
      <header class="section-header compact">
        <h3 class="section-title">{{
          $t('app.logrotate.form.summary.title')
        }}</h3>
        <p class="section-desc">{{ $t('app.logrotate.form.summary.desc') }}</p>
      </header>

      <div class="summary-grid">
        <div class="summary-item">
          <span class="summary-label">{{ $t('app.logrotate.form.name') }}</span>
          <code class="summary-value">{{ formData.name || '-' }}</code>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{
            $t('app.logrotate.form.category')
          }}</span>
          <code class="summary-value">{{ formData.category || '-' }}</code>
        </div>
        <div class="summary-item full">
          <span class="summary-label">{{ $t('app.logrotate.form.path') }}</span>
          <code class="summary-value">{{ formData.path || '-' }}</code>
        </div>
      </div>

      <div class="summary-tags">
        <a-tag color="arcoblue">{{
          $t(`app.logrotate.frequency.${formData.frequency}`)
        }}</a-tag>
        <a-tag color="blue"
          >{{ $t('app.logrotate.form.count') }} {{ formData.count }}</a-tag
        >
        <a-tag :color="formData.compress ? 'green' : 'gray'">
          {{ formData.compress ? 'compress' : 'nocompress' }}
        </a-tag>
        <a-tag :color="formData.delayCompress ? 'green' : 'gray'">
          {{ formData.delayCompress ? 'delaycompress' : 'nodelaycompress' }}
        </a-tag>
        <a-tag :color="formData.missingOk ? 'green' : 'gray'">
          {{ formData.missingOk ? 'missingok' : 'nomissingok' }}
        </a-tag>
        <a-tag :color="formData.notIfEmpty ? 'green' : 'gray'">
          {{ formData.notIfEmpty ? 'notifempty' : 'ifempty' }}
        </a-tag>
      </div>
    </section>

    <section class="param-section">
      <header class="section-header">
        <h3 class="section-title">{{
          $t('app.logrotate.form.section.basic')
        }}</h3>
        <p class="section-desc">{{
          $t('app.logrotate.form.section.basic_desc')
        }}</p>
      </header>

      <a-form-item field="name" class="param-item">
        <div class="param-head">
          <span class="param-label">{{ $t('app.logrotate.form.name') }}</span>
          <code class="param-key">name</code>
        </div>
        <div class="param-value">
          <a-input
            :model-value="formData.name"
            :placeholder="$t('app.logrotate.form.name_placeholder')"
            :disabled="isEdit"
            @update:model-value="(value: string) => updateFormData('name', value)"
          />
        </div>
        <p class="param-help">{{ $t('app.logrotate.form.name_help') }}</p>
      </a-form-item>

      <a-form-item field="category" class="param-item">
        <div class="param-head">
          <span class="param-label">{{
            $t('app.logrotate.form.category')
          }}</span>
          <code class="param-key">category</code>
        </div>
        <div class="param-value">
          <a-select
            :model-value="formData.category"
            :placeholder="$t('app.logrotate.form.category_placeholder')"
            allow-search
            allow-create
            :disabled="isSystemType || isEdit"
            :loading="categoryLoading"
            @visible-change="handleCategoryVisibleChange"
            @update:model-value="
              (value: string) => updateFormData('category', value)
            "
          >
            <a-option
              v-for="option in categoryOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-option>
          </a-select>
        </div>
        <p class="param-help">{{ $t('app.logrotate.form.category_help') }}</p>
      </a-form-item>

      <a-form-item field="path" class="param-item">
        <div class="param-head">
          <span class="param-label">{{ $t('app.logrotate.form.path') }}</span>
          <code class="param-key">/path/to/logfile</code>
        </div>
        <div class="param-value">
          <file-selector
            :model-value="formData.path"
            type="file"
            :host="hostId"
            :placeholder="$t('app.logrotate.form.path_placeholder')"
            @update:model-value="(value: string) => updateFormData('path', value)"
          />
        </div>
        <p class="param-help">{{ $t('app.logrotate.form.path_help') }}</p>
      </a-form-item>
    </section>

    <section class="param-section">
      <header class="section-header">
        <h3 class="section-title">
          {{ $t('app.logrotate.form.section.strategy') }}
        </h3>
        <p class="section-desc">
          {{ $t('app.logrotate.form.section.strategy_desc') }}
        </p>
      </header>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item field="frequency" class="param-item">
            <div class="param-head">
              <span class="param-label">{{
                $t('app.logrotate.form.frequency')
              }}</span>
              <code class="param-key">daily | weekly | monthly | yearly</code>
            </div>
            <div class="param-value">
              <a-select
                :model-value="formData.frequency"
                :placeholder="$t('app.logrotate.form.frequency_placeholder')"
                @update:model-value="
                  (value: string) => updateFormData('frequency', value)
                "
              >
                <a-option
                  v-for="freq in frequencyOptions"
                  :key="freq.value"
                  :value="freq.value"
                >
                  {{ freq.label }}
                </a-option>
              </a-select>
            </div>
            <p class="param-help">{{
              $t('app.logrotate.form.frequency_help')
            }}</p>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item field="count" class="param-item">
            <div class="param-head">
              <span class="param-label">{{
                $t('app.logrotate.form.count')
              }}</span>
              <code class="param-key">rotate &lt;N&gt;</code>
            </div>
            <div class="param-value">
              <a-input-number
                :model-value="formData.count"
                :placeholder="$t('app.logrotate.form.count_placeholder')"
                :min="1"
                :precision="0"
                @update:model-value="
                  (value: number | undefined) => updateFormData('count', value)
                "
              />
            </div>
            <p class="param-help">{{ $t('app.logrotate.form.count_help') }}</p>
          </a-form-item>
        </a-col>
      </a-row>

      <a-form-item class="param-item">
        <div class="param-head">
          <span class="param-label">{{
            $t('app.logrotate.form.rotate_options')
          }}</span>
          <code class="param-key">compress / missingok / notifempty</code>
        </div>
        <div class="param-value">
          <div class="option-grid">
            <div class="option-item">
              <a-checkbox
                :model-value="formData.compress"
                @update:model-value="handleCompressChange"
              >
                {{ $t('app.logrotate.form.compress') }}
              </a-checkbox>
              <p class="option-help">{{
                $t('app.logrotate.form.compress_help')
              }}</p>
            </div>
            <div class="option-item">
              <a-checkbox
                :model-value="formData.delayCompress"
                :disabled="!formData.compress"
                @update:model-value="
                  (value: boolean) => updateFormData('delayCompress', value)
                "
              >
                {{ $t('app.logrotate.form.delay_compress') }}
              </a-checkbox>
              <p class="option-help">
                {{ $t('app.logrotate.form.delay_compress_help') }}
              </p>
            </div>
            <div class="option-item">
              <a-checkbox
                :model-value="formData.missingOk"
                @update:model-value="
                  (value: boolean) => updateFormData('missingOk', value)
                "
              >
                {{ $t('app.logrotate.form.missing_ok') }}
              </a-checkbox>
              <p class="option-help">{{
                $t('app.logrotate.form.missing_ok_help')
              }}</p>
            </div>
            <div class="option-item">
              <a-checkbox
                :model-value="formData.notIfEmpty"
                @update:model-value="
                  (value: boolean) => updateFormData('notIfEmpty', value)
                "
              >
                {{ $t('app.logrotate.form.not_if_empty') }}
              </a-checkbox>
              <p class="option-help">
                {{ $t('app.logrotate.form.not_if_empty_help') }}
              </p>
            </div>
          </div>
        </div>
      </a-form-item>
    </section>

    <a-collapse :default-active-key="['permission']" class="advanced-collapse">
      <a-collapse-item
        key="permission"
        :header="$t('app.logrotate.form.advanced.permission')"
      >
        <section class="param-section nested">
          <a-form-item class="param-item">
            <div class="param-head">
              <span class="param-label">{{
                $t('app.logrotate.form.create')
              }}</span>
              <code class="param-key">create 0644 root root</code>
            </div>
            <div class="param-value">
              <PermissionInput
                :model-value="formData.create"
                @update:model-value="
                  (value: string) => updateFormData('create', value)
                "
              />
            </div>
            <p class="param-help">{{ $t('app.logrotate.form.create_help') }}</p>
          </a-form-item>
        </section>
      </a-collapse-item>

      <a-collapse-item
        key="script"
        :header="$t('app.logrotate.form.advanced.script')"
      >
        <section class="param-section nested">
          <a-form-item field="preRotate" class="param-item">
            <div class="param-head">
              <span class="param-label">{{
                $t('app.logrotate.form.pre_rotate')
              }}</span>
              <code class="param-key">prerotate ... endscript</code>
            </div>
            <div class="param-value">
              <div class="script-toolbar">
                <a-space size="small" wrap>
                  <a-select
                    :model-value="preRotateTemplate"
                    :placeholder="
                      $t('app.logrotate.form.script_tpl.select_placeholder')
                    "
                    style="width: 220px"
                    size="small"
                    @update:model-value="
                      (value: ScriptTemplateKey) => (preRotateTemplate = value)
                    "
                  >
                    <a-option
                      v-for="item in preScriptTemplateOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </a-option>
                  </a-select>
                  <a-button
                    size="small"
                    type="outline"
                    :disabled="!preRotateTemplate"
                    @click="
                      insertScriptTemplate('preRotate', preRotateTemplate)
                    "
                  >
                    {{ $t('app.logrotate.form.script_tpl.insert') }}
                  </a-button>
                  <a-button
                    size="small"
                    type="text"
                    :disabled="!formData.preRotate"
                    @click="updateFormData('preRotate', '')"
                  >
                    {{ $t('app.logrotate.form.clear') }}
                  </a-button>
                </a-space>
              </div>
              <a-textarea
                class="script-editor"
                :model-value="formData.preRotate"
                :placeholder="$t('app.logrotate.form.pre_rotate_placeholder')"
                :auto-size="{ minRows: 4, maxRows: 12 }"
                @update:model-value="(value: string) => updateFormData('preRotate', value)"
              />
            </div>
            <p class="param-help">{{
              $t('app.logrotate.form.pre_rotate_help')
            }}</p>
          </a-form-item>

          <a-form-item field="postRotate" class="param-item">
            <div class="param-head">
              <span class="param-label">{{
                $t('app.logrotate.form.post_rotate')
              }}</span>
              <code class="param-key">postrotate ... endscript</code>
            </div>
            <div class="param-value">
              <div class="script-toolbar">
                <a-space size="small" wrap>
                  <a-select
                    :model-value="postRotateTemplate"
                    :placeholder="
                      $t('app.logrotate.form.script_tpl.select_placeholder')
                    "
                    style="width: 220px"
                    size="small"
                    @update:model-value="
                      (value: ScriptTemplateKey) => (postRotateTemplate = value)
                    "
                  >
                    <a-option
                      v-for="item in postScriptTemplateOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </a-option>
                  </a-select>
                  <a-button
                    size="small"
                    type="outline"
                    :disabled="!postRotateTemplate"
                    @click="
                      insertScriptTemplate('postRotate', postRotateTemplate)
                    "
                  >
                    {{ $t('app.logrotate.form.script_tpl.insert') }}
                  </a-button>
                  <a-button
                    size="small"
                    type="text"
                    :disabled="!formData.postRotate"
                    @click="updateFormData('postRotate', '')"
                  >
                    {{ $t('app.logrotate.form.clear') }}
                  </a-button>
                </a-space>
              </div>
              <a-textarea
                class="script-editor"
                :model-value="formData.postRotate"
                :placeholder="$t('app.logrotate.form.post_rotate_placeholder')"
                :auto-size="{ minRows: 4, maxRows: 12 }"
                @update:model-value="
                  (value: string) => updateFormData('postRotate', value)
                "
              />
            </div>
            <p class="param-help">{{
              $t('app.logrotate.form.post_rotate_help')
            }}</p>
          </a-form-item>
        </section>
      </a-collapse-item>
    </a-collapse>

    <section class="preview-section">
      <div class="preview-head">
        <h4>{{ $t('app.logrotate.form.preview.title') }}</h4>
        <span>{{ $t('app.logrotate.form.preview.desc') }}</span>
      </div>
      <pre class="preview-content">{{ generatedPreview }}</pre>
    </section>
  </a-form>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { LOGROTATE_TYPE } from '@/config/enum';
  import FileSelector from '@/components/file/file-selector/index.vue';
  import PermissionInput from './permission-input.vue';
  import type { FormData, SelectOption } from './types';
  import { generateLogrotateContentFromForm } from '../../utils/content';

  interface Props {
    formData: FormData;
    formRules: Record<string, any>;
    frequencyOptions: SelectOption[];
    categoryOptions: SelectOption[];
    categoryLoading: boolean;
    isEdit: boolean;
    currentType: LOGROTATE_TYPE;
    hostId?: number;
  }

  type ScriptTemplateKey =
    | 'sharedscripts'
    | 'reloadNginx'
    | 'reloadRsyslog'
    | '';

  const emit = defineEmits<{
    updateFormData: [field: keyof FormData, value: any];
    fetchCategories: [];
  }>();

  const props = defineProps<Props>();
  const { t } = useI18n();
  const formRef = ref();
  const isSystemType = computed(
    () => props.currentType === LOGROTATE_TYPE.System
  );

  const preRotateTemplate = ref<ScriptTemplateKey>('');
  const postRotateTemplate = ref<ScriptTemplateKey>('');

  const scriptTemplateMap: Record<Exclude<ScriptTemplateKey, ''>, string> = {
    sharedscripts: 'sharedscripts',
    reloadNginx: 'systemctl reload nginx >/dev/null 2>&1 || true',
    reloadRsyslog: 'systemctl reload rsyslog >/dev/null 2>&1 || true',
  };

  const preScriptTemplateOptions = computed(() => [
    {
      label: t('app.logrotate.form.script_tpl.sharedscripts'),
      value: 'sharedscripts' as ScriptTemplateKey,
    },
    {
      label: t('app.logrotate.form.script_tpl.reload_nginx'),
      value: 'reloadNginx' as ScriptTemplateKey,
    },
  ]);

  const postScriptTemplateOptions = computed(() => [
    {
      label: t('app.logrotate.form.script_tpl.sharedscripts'),
      value: 'sharedscripts' as ScriptTemplateKey,
    },
    {
      label: t('app.logrotate.form.script_tpl.reload_nginx'),
      value: 'reloadNginx' as ScriptTemplateKey,
    },
    {
      label: t('app.logrotate.form.script_tpl.reload_rsyslog'),
      value: 'reloadRsyslog' as ScriptTemplateKey,
    },
  ]);

  const generatedPreview = computed(() => {
    if (!props.formData.path?.trim()) {
      return t('app.logrotate.form.preview.empty');
    }
    return generateLogrotateContentFromForm(props.formData, {
      includeHeader: true,
      indent: '  ',
    });
  });

  const updateFormData = (field: keyof FormData, value: any) => {
    emit('updateFormData', field, value);
  };

  const handleCategoryVisibleChange = (visible: boolean) => {
    if (!visible || isSystemType.value) {
      return;
    }
    emit('fetchCategories');
  };

  const handleCompressChange = (value: boolean) => {
    updateFormData('compress', value);
    if (!value && props.formData.delayCompress) {
      updateFormData('delayCompress', false);
    }
  };

  const insertScriptTemplate = (
    field: 'preRotate' | 'postRotate',
    key: ScriptTemplateKey
  ) => {
    if (!key) return;

    const template = scriptTemplateMap[key as Exclude<ScriptTemplateKey, ''>];
    if (!template) return;

    const currentValue = (props.formData[field] || '').trim();
    const nextValue = currentValue ? `${currentValue}\n${template}` : template;
    updateFormData(field, nextValue);
  };

  defineExpose({
    validate: () => formRef.value?.validate(),
    clearValidate: () => formRef.value?.clearValidate(),
    resetFields: () => formRef.value?.resetFields(),
  });
</script>

<style scoped lang="less">
  :deep(.arco-form-item-label-col) {
    display: none;
  }

  :deep(.arco-form-item-layout-vertical .arco-form-item-wrapper-col) {
    margin-top: 0;
  }

  .summary-section,
  .param-section,
  .preview-section {
    padding: 14px;
    margin-bottom: 20px;
    background: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }

  .section-header {
    margin-bottom: 10px;
  }

  .section-header.compact {
    margin-bottom: 12px;
  }

  .section-title {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .section-desc {
    margin: 4px 0 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--color-text-3);
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(220px, 1fr));
    gap: 10px;
    margin-bottom: 10px;
  }

  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .summary-item.full {
    grid-column: 1 / -1;
  }

  .summary-label {
    font-size: 12px;
    color: var(--color-text-3);
  }

  .summary-value {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    word-break: break-all;
    white-space: pre-wrap;
  }

  .summary-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .advanced-collapse {
    margin-bottom: 20px;
  }

  .advanced-collapse :deep(.arco-collapse-item-header-title) {
    font-weight: 600;
  }

  .param-section.nested {
    padding: 0;
    margin: 0;
    background: transparent;
    border: none;
  }

  .param-item {
    padding: 12px;
    margin-bottom: 12px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .param-item:last-child {
    margin-bottom: 0;
  }

  .param-head {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .param-label {
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .param-key {
    padding: 2px 6px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 11px;
    color: var(--color-text-2);
    background: var(--color-fill-2);
    border-radius: 4px;
  }

  .param-value {
    width: 100%;
  }

  .param-help {
    margin: 8px 0 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--color-text-3);
  }

  .option-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(220px, 1fr));
    gap: 10px 14px;
    padding: 10px;
    background: var(--color-fill-1);
    border: 1px dashed var(--color-border-2);
    border-radius: 6px;
  }

  .option-item {
    padding: 8px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .option-help {
    margin: 6px 0 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--color-text-3);
  }

  .script-toolbar {
    margin-bottom: 8px;
  }

  :deep(.script-editor textarea) {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 13px;
    line-height: 1.5;
  }

  .preview-section {
    margin-bottom: 0;
  }

  .preview-head {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    color: var(--color-text-3);
  }

  .preview-head h4 {
    margin: 0;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .preview-head span {
    font-size: 12px;
  }

  .preview-content {
    min-height: 140px;
    max-height: 260px;
    padding: 10px;
    margin: 0;
    overflow: auto;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.5;
    white-space: pre-wrap;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  :deep(.arco-checkbox-checked .arco-checkbox-icon) {
    background-color: var(--idblue-6) !important;
    border-color: var(--idblue-6) !important;
  }

  :deep(.arco-checkbox-checked .arco-checkbox-icon .arco-checkbox-icon-check) {
    color: var(--idb-brand-text) !important;
  }

  :deep(.arco-checkbox:not(.arco-checkbox-disabled):hover .arco-checkbox-icon) {
    border-color: var(--idblue-6) !important;
  }

  :deep(
      .arco-checkbox:not(.arco-checkbox-disabled).arco-checkbox-focus
        .arco-checkbox-icon
    ) {
    border-color: var(--idblue-6) !important;
    box-shadow: 0 0 0 2px var(--idblue-1) !important;
  }

  @media (width <= 768px) {
    .param-head,
    .preview-head {
      flex-direction: column;
      gap: 4px;
      align-items: flex-start;
    }
    .summary-grid,
    .option-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
