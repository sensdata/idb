<template>
  <a-spin :loading="loading">
    <div class="box">
      <div class="line">
        <div class="col1">
          <div class="main-label">
            {{ $t('app.sysinfo.config.max_watch_files') }}
          </div>
          <div class="extra-label"> fs.inotify.max_user_watches </div>
        </div>
        <div class="col2">
          <a-input-number
            v-model="formState.max_watch_files"
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
            {{ $t('app.sysinfo.config.max_open_files') }}
          </div>
          <div class="extra-label"> DefaultLimitNOFILE </div>
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

      <div class="line no-border">
        <div class="col1"></div>
        <div class="col2">
          <a-button type="primary" :loading="saving" @click="handleSave">
            {{ $t('common.save') }}
          </a-button>
        </div>
      </div>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
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

  // 表单数据
  const formState = reactive<UpdateSettingsParams>({
    max_open_files: 1024,
    max_watch_files: 8192,
  });

  // 原始数据，用于取消编辑时恢复
  const originalData = ref<UpdateSettingsParams>({
    max_open_files: 1024,
    max_watch_files: 8192,
  });

  // 获取系统配置数据
  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoSettingsApi();

      // 更新表单数据和原始数据
      Object.assign(formState, res);
      originalData.value = { ...res };
    } catch (err: any) {
      Message.error(err.message || 'Failed to fetch system settings');
    } finally {
      setLoading(false);
    }
  };

  // 保存配置
  const handleSave = async () => {
    try {
      saving.value = true;

      await updateSysInfoSettingsApi({
        max_open_files: formState.max_open_files,
        max_watch_files: formState.max_watch_files,
      });

      // 更新原始数据
      originalData.value = { ...formState };

      Message.success(t('app.sysinfo.config.save_success'));
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.config.save_failed'));
    } finally {
      saving.value = false;
    }
  };

  onMounted(() => {
    fetchData();
  });
</script>

<style scoped lang="less">
  .box {
    width: 940px;
    margin: 0 auto;
    padding: 0 16px;
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

  .no-border {
    border-bottom: none;
  }

  .col1 {
    width: 220px;
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
  }

  .col2 {
    flex: 1;
    min-width: 200px;
    font-size: 14px;
  }
</style>
