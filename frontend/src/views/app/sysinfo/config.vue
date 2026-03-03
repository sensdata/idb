<template>
  <a-spin :loading="loading">
    <div class="box">
      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.template.title') }}
          </div>
          <div class="hint">{{ $t('app.sysinfo.config.template.desc') }}</div>
        </div>
        <div class="col2">
          <div class="template-actions">
            <a-space wrap>
              <a-button
                :type="activeTemplate === 'general' ? 'primary' : 'outline'"
                @click="applyTemplate('general')"
              >
                {{ $t('app.sysinfo.config.template.general') }}
              </a-button>
              <a-button
                :type="activeTemplate === 'container' ? 'primary' : 'outline'"
                @click="applyTemplate('container')"
              >
                {{ $t('app.sysinfo.config.template.container') }}
              </a-button>
              <a-button
                :type="activeTemplate === 'database' ? 'primary' : 'outline'"
                @click="applyTemplate('database')"
              >
                {{ $t('app.sysinfo.config.template.database') }}
              </a-button>
              <a-button
                :type="
                  activeTemplate === 'high_concurrency' ? 'primary' : 'outline'
                "
                @click="applyTemplate('high_concurrency')"
              >
                {{ $t('app.sysinfo.config.template.high_concurrency') }}
              </a-button>
              <a-button @click="restoreCurrentSettings">
                {{ $t('app.sysinfo.config.template.reset_current') }}
              </a-button>
            </a-space>
            <div class="save-actions">
              <span v-if="hasUnsavedChanges" class="unsaved-tip">
                {{ $t('app.sysinfo.config.unsaved_changes') }}
              </span>
              <a-button type="primary" :loading="saving" @click="handleSave">
                {{ $t('common.save') }}
              </a-button>
            </div>
          </div>
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.max_watch_files') }}
          </div>
          <div class="extra-label"> fs.inotify.max_user_watches </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.max_watch_files')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_watch_files"
            class="w-60"
            :min="8192"
            :step="1024"
          />
          <span class="ml-2">{{ $t('app.sysinfo.config.files') }}</span>
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.max_watch_instances') }}
          </div>
          <div class="extra-label"> fs.inotify.max_user_instances </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.max_watch_instances')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_watch_instances"
            class="w-60"
            :min="128"
            :step="64"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.max_queued_events') }}
          </div>
          <div class="extra-label"> fs.inotify.max_queued_events </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.max_queued_events')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_queued_events"
            class="w-60"
            :min="16384"
            :step="1024"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.max_open_files') }}
          </div>
          <div class="extra-label"> DefaultLimitNOFILE </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.max_open_files')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_open_files"
            class="w-60"
            :min="1024"
            :step="1024"
          />
          <span class="ml-2">{{ $t('app.sysinfo.config.files') }}</span>
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.file_max') }}
          </div>
          <div class="extra-label"> fs.file-max </div>
          <div class="hint">{{ $t('app.sysinfo.config.tip.file_max') }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.file_max"
            class="w-60"
            :min="65535"
            :step="1024"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{
            $t('app.sysinfo.config.swappiness')
          }}</div>
          <div class="extra-label"> vm.swappiness </div>
          <div class="hint">{{ $t('app.sysinfo.config.tip.swappiness') }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.swappiness"
            class="w-60"
            :min="0"
            :max="100"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{
            $t('app.sysinfo.config.max_map_count')
          }}</div>
          <div class="extra-label"> vm.max_map_count </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.max_map_count')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_map_count"
            class="w-60"
            :min="65530"
            :step="1024"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{ $t('app.sysinfo.config.somaxconn') }}</div>
          <div class="extra-label"> net.core.somaxconn </div>
          <div class="hint">{{ $t('app.sysinfo.config.tip.somaxconn') }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.somaxconn"
            class="w-60"
            :min="128"
            :step="128"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.tcp_max_syn_backlog') }}
          </div>
          <div class="extra-label"> net.ipv4.tcp_max_syn_backlog </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.tcp_max_syn_backlog')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.tcp_max_syn_backlog"
            class="w-60"
            :min="128"
            :step="128"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{ $t('app.sysinfo.config.pid_max') }}</div>
          <div class="extra-label"> kernel.pid_max </div>
          <div class="hint">{{ $t('app.sysinfo.config.tip.pid_max') }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.pid_max"
            class="w-60"
            :min="32768"
            :step="1024"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{
            $t('app.sysinfo.config.overcommit_memory')
          }}</div>
          <div class="extra-label"> vm.overcommit_memory </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.overcommit_memory')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.overcommit_memory"
            class="w-60"
            :min="0"
            :max="2"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{
            $t('app.sysinfo.config.overcommit_ratio')
          }}</div>
          <div class="extra-label"> vm.overcommit_ratio </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.overcommit_ratio')
          }}</div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.overcommit_ratio"
            class="w-60"
            :min="0"
            :max="100"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">
          <div class="main-label">{{
            $t('app.sysinfo.config.transparent_huge_page')
          }}</div>
          <div class="extra-label">
            /sys/kernel/mm/transparent_hugepage/enabled
          </div>
          <div class="hint">{{
            $t('app.sysinfo.config.tip.transparent_huge_page')
          }}</div>
        </div>
        <div class="col2">
          <a-select v-model="formState.transparent_huge_page" class="w-60">
            <a-option value="always">always</a-option>
            <a-option value="madvise">madvise</a-option>
            <a-option value="never">never</a-option>
          </a-select>
        </div>
      </div>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    getSysInfoSettingsApi,
    updateSysInfoSettingsApi,
    UpdateSettingsParams,
  } from '@/api/sysinfo';
  import useLoading from '@/composables/loading';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const saving = ref(false);
  type TemplateKey = 'general' | 'container' | 'database' | 'high_concurrency';
  const activeTemplate = ref<'custom' | TemplateKey>('custom');

  const getDefaultFormState = (): UpdateSettingsParams => ({
    file_max: 2097152,
    max_map_count: 262144,
    max_open_files: 1048576,
    max_queued_events: 32768,
    max_watch_files: 524288,
    max_watch_instances: 1024,
    overcommit_memory: 0,
    overcommit_ratio: 50,
    pid_max: 262144,
    somaxconn: 4096,
    swappiness: 10,
    tcp_max_syn_backlog: 4096,
    transparent_huge_page: 'madvise',
  });

  const formState = reactive<UpdateSettingsParams>(getDefaultFormState());
  const originalData = ref<UpdateSettingsParams>(getDefaultFormState());
  const settingKeys: Array<keyof UpdateSettingsParams> = [
    'file_max',
    'max_map_count',
    'max_open_files',
    'max_queued_events',
    'max_watch_files',
    'max_watch_instances',
    'overcommit_memory',
    'overcommit_ratio',
    'pid_max',
    'somaxconn',
    'swappiness',
    'tcp_max_syn_backlog',
    'transparent_huge_page',
  ];
  const hasUnsavedChanges = computed(() => {
    return settingKeys.some(
      (key) => formState[key] !== originalData.value[key]
    );
  });

  const SETTINGS_TEMPLATES: Record<TemplateKey, UpdateSettingsParams> = {
    general: {
      file_max: 2097152,
      max_map_count: 262144,
      max_open_files: 1048576,
      max_queued_events: 32768,
      max_watch_files: 524288,
      max_watch_instances: 1024,
      overcommit_memory: 0,
      overcommit_ratio: 50,
      pid_max: 262144,
      somaxconn: 4096,
      swappiness: 10,
      tcp_max_syn_backlog: 4096,
      transparent_huge_page: 'madvise',
    },
    container: {
      file_max: 2097152,
      max_map_count: 262144,
      max_open_files: 1048576,
      max_queued_events: 65536,
      max_watch_files: 1048576,
      max_watch_instances: 2048,
      overcommit_memory: 1,
      overcommit_ratio: 80,
      pid_max: 262144,
      somaxconn: 4096,
      swappiness: 10,
      tcp_max_syn_backlog: 8192,
      transparent_huge_page: 'madvise',
    },
    database: {
      file_max: 2097152,
      max_map_count: 262144,
      max_open_files: 1048576,
      max_queued_events: 32768,
      max_watch_files: 262144,
      max_watch_instances: 1024,
      overcommit_memory: 2,
      overcommit_ratio: 90,
      pid_max: 262144,
      somaxconn: 1024,
      swappiness: 1,
      tcp_max_syn_backlog: 2048,
      transparent_huge_page: 'never',
    },
    high_concurrency: {
      file_max: 4194304,
      max_map_count: 262144,
      max_open_files: 2097152,
      max_queued_events: 65536,
      max_watch_files: 524288,
      max_watch_instances: 2048,
      overcommit_memory: 1,
      overcommit_ratio: 80,
      pid_max: 524288,
      somaxconn: 16384,
      swappiness: 10,
      tcp_max_syn_backlog: 16384,
      transparent_huge_page: 'madvise',
    },
  };

  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoSettingsApi();
      Object.assign(formState, res);
      originalData.value = { ...res };
      activeTemplate.value = 'custom';
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.config.fetch_failed'));
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      saving.value = true;
      await updateSysInfoSettingsApi({ ...formState });
      originalData.value = { ...formState };
      Message.success(t('app.sysinfo.config.save_success'));
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.config.save_failed'));
    } finally {
      saving.value = false;
    }
  };

  const applyTemplate = (templateKey: TemplateKey) => {
    Object.assign(formState, SETTINGS_TEMPLATES[templateKey]);
    activeTemplate.value = templateKey;
    Message.success(
      t('app.sysinfo.config.template.apply_success', {
        name: t(`app.sysinfo.config.template.${templateKey}`),
      })
    );
  };

  const restoreCurrentSettings = () => {
    Object.assign(formState, originalData.value);
    activeTemplate.value = 'custom';
  };

  onMounted(() => {
    fetchData();
  });
</script>

<style scoped lang="less">
  .box {
    width: 940px;
    padding: 0 16px;
    margin: 0 auto;
    border: 1px solid var(--color-border-2);
  }

  .line {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    width: 100%;
    padding: 16px 0;
    border-bottom: 1px solid var(--color-border);
    &:last-child {
      border-bottom: none;
    }
  }

  .col1 {
    width: 320px;
    margin-right: 20px;
    font-size: 14px;
    text-align: right;
    .main-label {
      margin-bottom: 4px;
      color: var(--color-text-2);
    }
    .extra-label {
      padding-right: 10px;
      color: var(--color-text-3);
    }
    .hint {
      padding-right: 10px;
      margin-top: 4px;
      font-size: 12px;
      line-height: 16px;
      color: var(--color-text-3);
    }
  }

  .col2 {
    flex: 1;
    min-width: 200px;
    font-size: 14px;
  }

  .template-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .save-actions {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .unsaved-tip {
    font-size: 12px;
    color: rgb(var(--danger-6));
  }
</style>
