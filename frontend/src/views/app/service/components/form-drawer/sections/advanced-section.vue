<template>
  <div :class="[styles.form_section, styles.advanced_section_container]">
    <div :class="styles.section_header" @click="toggleAdvanced">
      <h3 :class="styles.section_title">
        {{ $t('app.service.form.advanced.title') }}
      </h3>
      <icon-down :class="{ [styles.icon_rotate]: showAdvanced }" />
    </div>

    <div v-show="showAdvanced" :class="styles.advanced_section">
      <h4 :class="styles.subsection_title">
        {{ $t('app.service.form.section.lifecycle') }}
      </h4>

      <div :class="styles.form_row">
        <a-form-item
          field="execStop"
          :label="$t('app.service.form.field.exec_stop')"
          :class="styles.form_item_half"
        >
          <a-input
            :model-value="formModel.execStop"
            :placeholder="$t('app.service.form.placeholder.exec_stop')"
            allow-clear
            @update:model-value="updateField('execStop', $event)"
          >
            <template #prefix>
              <icon-stop />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          field="execReload"
          :label="$t('app.service.form.field.exec_reload')"
          :class="styles.form_item_half"
        >
          <a-input
            :model-value="formModel.execReload"
            :placeholder="$t('app.service.form.placeholder.exec_reload')"
            allow-clear
            @update:model-value="updateField('execReload', $event)"
          >
            <template #prefix>
              <icon-refresh />
            </template>
          </a-input>
        </a-form-item>
      </div>

      <h4 :class="styles.subsection_title">
        {{ $t('app.service.form.section.restart') }}
      </h4>

      <div :class="styles.form_row">
        <a-form-item
          field="restart"
          :label="$t('app.service.form.field.restart')"
          :class="styles.form_item_half"
        >
          <a-select
            :model-value="formModel.restart"
            :placeholder="$t('app.service.form.field.restart')"
            @update:model-value="updateField('restart', $event)"
          >
            <a-option value="no">No</a-option>
            <a-option value="on-success">On Success</a-option>
            <a-option value="on-failure">On Failure</a-option>
            <a-option value="on-abnormal">On Abnormal</a-option>
            <a-option value="on-watchdog">On Watchdog</a-option>
            <a-option value="on-abort">On Abort</a-option>
            <a-option value="always">Always</a-option>
          </a-select>
        </a-form-item>

        <a-form-item
          field="restartSec"
          :label="$t('app.service.form.field.restart_sec')"
          :class="styles.form_item_half"
        >
          <a-input-number
            :model-value="formModel.restartSec"
            :placeholder="$t('app.service.form.placeholder.restart_sec')"
            :min="0"
            :precision="0"
            mode="button"
            :class="styles.full_width"
            @update:model-value="updateField('restartSec', $event)"
          />
          <template #suffix>
            <span :class="styles.input_suffix">
              {{ $t('app.service.form.unit.seconds') }}
            </span>
          </template>
        </a-form-item>
      </div>

      <h4 :class="styles.subsection_title">
        {{ $t('app.service.form.section.timeouts') }}
      </h4>

      <div :class="styles.form_row">
        <a-form-item
          field="timeoutStartSec"
          :label="$t('app.service.form.field.timeout_start_sec')"
          :class="styles.form_item_half"
        >
          <a-input-number
            :model-value="formModel.timeoutStartSec"
            :placeholder="$t('app.service.form.placeholder.timeout_start_sec')"
            :min="0"
            :precision="0"
            mode="button"
            :class="styles.full_width"
            @update:model-value="updateField('timeoutStartSec', $event)"
          />
          <template #suffix>
            <span :class="styles.input_suffix">
              {{ $t('app.service.form.unit.seconds') }}
            </span>
          </template>
        </a-form-item>

        <a-form-item
          field="timeoutStopSec"
          :label="$t('app.service.form.field.timeout_stop_sec')"
          :class="styles.form_item_half"
        >
          <a-input-number
            :model-value="formModel.timeoutStopSec"
            :placeholder="$t('app.service.form.placeholder.timeout_stop_sec')"
            :min="0"
            :precision="0"
            mode="button"
            :class="styles.full_width"
            @update:model-value="updateField('timeoutStopSec', $event)"
          />
          <template #suffix>
            <span :class="styles.input_suffix">
              {{ $t('app.service.form.unit.seconds') }}
            </span>
          </template>
        </a-form-item>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import type { ParsedServiceConfig } from '../types';

  const props = defineProps<{
    formModel: ParsedServiceConfig;
    styles: Record<string, string>;
  }>();

  const emit = defineEmits<{
    'update:formModel': [value: ParsedServiceConfig];
  }>();

  const showAdvanced = ref(false);

  const toggleAdvanced = () => {
    showAdvanced.value = !showAdvanced.value;
  };

  const updateField = (field: string, value: unknown) => {
    const updatedModel = { ...props.formModel };
    (updatedModel as any)[field] = value;
    emit('update:formModel', updatedModel);
  };
</script>
