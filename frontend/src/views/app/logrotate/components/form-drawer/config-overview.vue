<template>
  <div class="config-overview">
    <a-alert :type="overviewAlertType" show-icon class="overview-alert">
      {{ $t(overviewTipKey) }}
    </a-alert>

    <section class="overview-section">
      <h3 class="section-title">{{ $t('app.logrotate.overview.basic') }}</h3>
      <div class="kv-grid">
        <div class="kv-item">
          <span class="kv-label">
            {{ $t('app.logrotate.form.name') }}
            <span class="required-star">*</span>
          </span>
          <code v-if="!editing" class="kv-value">{{
            formData.name || '-'
          }}</code>
          <a-input
            v-else
            :model-value="formData.name"
            :placeholder="$t('app.logrotate.form.name_placeholder')"
            :disabled="isReadonly || isEdit"
            @update:model-value="(value: string) => emitUpdate('name', value)"
          />
        </div>

        <div class="kv-item full">
          <span class="kv-label">
            {{ $t('app.logrotate.form.path') }}
            <span class="required-star">*</span>
          </span>
          <code v-if="!editing" class="kv-value">{{
            formData.path || '-'
          }}</code>
          <file-selector
            v-else
            :model-value="formData.path"
            type="file"
            :host="hostId"
            :placeholder="$t('app.logrotate.form.path_placeholder')"
            :disabled="isReadonly"
            @update:model-value="(value: string) => emitUpdate('path', value)"
          />
        </div>
      </div>
    </section>

    <section class="overview-section">
      <h3 class="section-title">{{ $t('app.logrotate.overview.strategy') }}</h3>
      <div class="strategy-row">
        <template v-if="!editing">
          <a-tag color="arcoblue">
            {{ $t(`app.logrotate.frequency.${formData.frequency}`) }}
          </a-tag>
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
        </template>

        <template v-else>
          <div class="directive-grid">
            <div class="directive-row">
              <code class="directive-key">frequency</code>
              <span class="required-star">*</span>
              <a-select
                class="directive-input"
                :style="{ width: '120px' }"
                :model-value="formData.frequency"
                size="small"
                :disabled="isReadonly"
                @update:model-value="
                  (value: string) => emitUpdate('frequency', value)
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

            <div class="directive-row">
              <code class="directive-key">rotate</code>
              <span class="required-star">*</span>
              <span class="directive-desc">{{
                $t('app.logrotate.form.count')
              }}</span>
              <a-input-number
                class="directive-input"
                :style="{ width: '120px' }"
                :model-value="formData.count"
                :min="1"
                :precision="0"
                size="small"
                :disabled="isReadonly"
                @update:model-value="
                  (value: number | undefined) => emitUpdate('count', value)
                "
              />
            </div>

            <div class="directive-row">
              <code class="directive-key">compress</code>
              <a-checkbox
                :model-value="formData.compress"
                :disabled="isReadonly"
                @update:model-value="handleCompressChange"
              >
                {{ $t('app.logrotate.form.compress') }}
              </a-checkbox>
            </div>

            <div class="directive-row">
              <code class="directive-key">delaycompress</code>
              <a-checkbox
                :model-value="formData.delayCompress"
                :disabled="isReadonly || !formData.compress"
                @update:model-value="
                  (value: boolean) => emitUpdate('delayCompress', value)
                "
              >
                {{ $t('app.logrotate.form.delay_compress') }}
              </a-checkbox>
            </div>

            <div class="directive-row">
              <code class="directive-key">missingok</code>
              <a-checkbox
                :model-value="formData.missingOk"
                :disabled="isReadonly"
                @update:model-value="(value: boolean) => emitUpdate('missingOk', value)"
              >
                {{ $t('app.logrotate.form.missing_ok') }}
              </a-checkbox>
            </div>

            <div class="directive-row">
              <code class="directive-key">notifempty</code>
              <a-checkbox
                :model-value="formData.notIfEmpty"
                :disabled="isReadonly"
                @update:model-value="(value: boolean) => emitUpdate('notIfEmpty', value)"
              >
                {{ $t('app.logrotate.form.not_if_empty') }}
              </a-checkbox>
            </div>
          </div>
        </template>
      </div>

      <p v-if="!editing" class="create-line">{{ formData.create || '-' }}</p>
      <PermissionInput
        v-else
        :model-value="formData.create"
        :disabled="isReadonly"
        @update:model-value="(value: string) => emitUpdate('create', value)"
      />
    </section>

    <section class="overview-section">
      <h3 class="section-title">{{ $t('app.logrotate.overview.script') }}</h3>
      <div class="script-grid">
        <div class="script-card">
          <div class="script-head">
            <div class="script-title">{{
              $t('app.logrotate.form.pre_rotate')
            }}</div>
            <div v-if="editing" class="script-toolbar">
              <a-space size="small" wrap>
                <a-select
                  :model-value="preRotateTemplate"
                  :placeholder="
                    $t('app.logrotate.form.script_tpl.select_placeholder')
                  "
                  style="width: 200px"
                  size="small"
                  :disabled="isReadonly"
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
                  :disabled="isReadonly || !preRotateTemplate"
                  @click="insertScriptTemplate('preRotate', preRotateTemplate)"
                >
                  {{ $t('app.logrotate.form.script_tpl.insert') }}
                </a-button>
              </a-space>
            </div>
          </div>
          <template v-if="!editing">
            <pre v-if="formData.preRotate" class="script-block">{{
              formData.preRotate
            }}</pre>
            <p v-else class="script-empty">{{
              $t('app.logrotate.overview.script_empty')
            }}</p>
          </template>
          <template v-else>
            <a-textarea
              class="script-editor"
              :model-value="formData.preRotate"
              :placeholder="$t('app.logrotate.form.pre_rotate_placeholder')"
              :auto-size="{ minRows: 4, maxRows: 12 }"
              :disabled="isReadonly"
              @update:model-value="(value: string) => emitUpdate('preRotate', value)"
            />
          </template>
        </div>

        <div class="script-card">
          <div class="script-head">
            <div class="script-title">{{
              $t('app.logrotate.form.post_rotate')
            }}</div>
            <div v-if="editing" class="script-toolbar">
              <a-space size="small" wrap>
                <a-select
                  :model-value="postRotateTemplate"
                  :placeholder="
                    $t('app.logrotate.form.script_tpl.select_placeholder')
                  "
                  style="width: 200px"
                  size="small"
                  :disabled="isReadonly"
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
                  :disabled="isReadonly || !postRotateTemplate"
                  @click="
                    insertScriptTemplate('postRotate', postRotateTemplate)
                  "
                >
                  {{ $t('app.logrotate.form.script_tpl.insert') }}
                </a-button>
              </a-space>
            </div>
          </div>
          <template v-if="!editing">
            <pre v-if="formData.postRotate" class="script-block">{{
              formData.postRotate
            }}</pre>
            <p v-else class="script-empty">{{
              $t('app.logrotate.overview.script_empty')
            }}</p>
          </template>
          <template v-else>
            <a-textarea
              class="script-editor"
              :model-value="formData.postRotate"
              :placeholder="$t('app.logrotate.form.post_rotate_placeholder')"
              :auto-size="{ minRows: 4, maxRows: 12 }"
              :disabled="isReadonly"
              @update:model-value="(value: string) => emitUpdate('postRotate', value)"
            />
          </template>
        </div>
      </div>
    </section>

    <section class="overview-section">
      <h3 class="section-title">{{
        $t('app.logrotate.overview.raw_preview')
      }}</h3>
      <pre class="raw-preview">{{ previewContent }}</pre>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { LOGROTATE_TYPE } from '@/config/enum';
  import FileSelector from '@/components/file/file-selector/index.vue';
  import PermissionInput from './permission-input.vue';
  import { generateLogrotateContentFromForm } from '../../utils/content';
  import type { FormData, SelectOption } from './types';

  interface Props {
    formData: FormData;
    rawContent: string;
    editing: boolean;
    isEdit: boolean;
    currentType: LOGROTATE_TYPE;
    frequencyOptions: SelectOption[];
    hostId?: number;
  }

  type ScriptTemplateKey =
    | 'sharedscripts'
    | 'reloadNginx'
    | 'reloadRsyslog'
    | '';

  const props = defineProps<Props>();
  const emit = defineEmits<{
    updateFormData: [field: keyof FormData, value: any];
  }>();

  const preRotateTemplate = ref<ScriptTemplateKey>('');
  const postRotateTemplate = ref<ScriptTemplateKey>('');

  const isSystemType = computed(
    () => props.currentType === LOGROTATE_TYPE.System
  );
  const isReadonly = computed(() => isSystemType.value);
  const editing = computed(() => props.editing);

  const scriptTemplateMap: Record<Exclude<ScriptTemplateKey, ''>, string> = {
    sharedscripts: 'sharedscripts',
    reloadNginx: 'systemctl reload nginx >/dev/null 2>&1 || true',
    reloadRsyslog: 'systemctl reload rsyslog >/dev/null 2>&1 || true',
  };

  const preScriptTemplateOptions = computed(() => [
    { value: 'sharedscripts', label: 'sharedscripts' },
    { value: 'reloadNginx', label: 'reload nginx' },
  ]);

  const postScriptTemplateOptions = computed(() => [
    { value: 'sharedscripts', label: 'sharedscripts' },
    { value: 'reloadNginx', label: 'reload nginx' },
    { value: 'reloadRsyslog', label: 'reload rsyslog' },
  ]);

  const previewContent = computed(() => {
    // 编辑态需要所见即所得，优先按当前结构化字段实时生成
    if (editing.value) {
      if (!props.formData.path?.trim()) {
        return '';
      }
      return generateLogrotateContentFromForm(props.formData, {
        includeHeader: true,
        indent: '  ',
      });
    }

    // 查看态优先展示后端原始内容
    if (props.rawContent?.trim()) {
      return props.rawContent;
    }

    if (!props.formData.path?.trim()) {
      return '';
    }

    return generateLogrotateContentFromForm(props.formData, {
      includeHeader: true,
      indent: '  ',
    });
  });

  const overviewTipKey = computed(() =>
    editing.value
      ? 'app.logrotate.overview.tip_edit'
      : 'app.logrotate.overview.tip_view'
  );

  const overviewAlertType = computed(() =>
    editing.value ? 'warning' : 'info'
  );

  const emitUpdate = (field: keyof FormData, value: any) => {
    emit('updateFormData', field, value);
  };

  const handleCompressChange = (value: boolean) => {
    emitUpdate('compress', value);
    if (!value && props.formData.delayCompress) {
      emitUpdate('delayCompress', false);
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
    emitUpdate(field, nextValue);
  };
</script>

<style scoped lang="less">
  .config-overview {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .overview-alert {
    margin-bottom: 4px;
  }

  .overview-section {
    padding: 14px;
    background: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }

  .section-title {
    margin: 0 0 10px;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .kv-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(220px, 1fr));
    gap: 10px;
  }

  .kv-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .full {
    grid-column: 1 / -1;
  }

  .kv-label {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    font-size: 13px;
    color: var(--color-text-3);
  }

  .required-star {
    font-size: 13px;
    font-weight: 600;
    line-height: 1;
    color: var(--red-6);
  }

  .kv-value {
    display: inline-block;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    color: var(--color-text-1);
    word-break: break-all;
    white-space: pre-wrap;
  }

  .strategy-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 10px;
  }

  .directive-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(220px, 1fr));
    gap: 8px;
    width: 100%;
  }

  .directive-row {
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 8px 10px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .directive-key {
    min-width: 90px;
    padding: 2px 6px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    color: var(--color-text-2);
    background: var(--color-fill-2);
    border-radius: 4px;
  }

  .directive-input {
    flex: 0 0 auto;
    width: 132px;
  }

  .directive-desc {
    font-size: 13px;
    color: var(--color-text-3);
    white-space: nowrap;
  }

  .create-line {
    padding: 8px 10px;
    margin: 0;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    background: var(--color-bg-2);
    border: 1px dashed var(--color-border-3);
    border-radius: 6px;
  }

  .script-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .script-card {
    padding: 10px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .script-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .script-head {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .script-block {
    max-height: 180px;
    padding: 8px;
    margin: 0;
    overflow: auto;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.5;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .script-empty {
    margin: 0;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .script-toolbar {
    margin-bottom: 0;
  }

  :deep(.script-editor textarea) {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 13px;
    line-height: 1.5;
  }

  .raw-preview {
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

  @media (width <= 768px) {
    .kv-grid,
    .script-grid,
    .directive-grid {
      grid-template-columns: 1fr;
    }
    .script-head {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
