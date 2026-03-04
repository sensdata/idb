<template>
  <a-tabs v-model:active-key="activeTab" lazy-load @change="handleTabChange">
    <template #extra>
      <a-button
        v-if="activeTab === CRONTAB_TYPE.System"
        size="small"
        @click="handleOpenConfigDirectory"
      >
        <template #icon>
          <icon-folder />
        </template>
        {{ $t('app.crontab.form.open_config_dir') }}
      </a-button>
    </template>
    <a-tab-pane
      :key="CRONTAB_TYPE.System"
      :title="$t('app.crontab.enum.type.system')"
    >
      <list
        :key="`list-${CRONTAB_TYPE.System}`"
        ref="systemListRef"
        :type="CRONTAB_TYPE.System"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="CRONTAB_TYPE.Local"
      :title="$t('app.crontab.enum.type.local')"
    >
      <list
        :key="`list-${CRONTAB_TYPE.Local}`"
        ref="localListRef"
        :type="CRONTAB_TYPE.Local"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="CRONTAB_TYPE.Global"
      :title="$t('app.crontab.enum.type.global')"
    >
      <list
        :key="`list-${CRONTAB_TYPE.Global}`"
        ref="globalListRef"
        :type="CRONTAB_TYPE.Global"
      />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { Message } from '@arco-design/web-vue';
  import { IconFolder } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { CRONTAB_TYPE } from '@/config/enum';
  import useCurrentHost from '@/composables/current-host';
  import { createFileRoute } from '@/utils/file-route';
  import { ref, onMounted, nextTick } from 'vue';
  import List from './list.vue';

  // 定义 List 组件实例类型
  interface ListInstance {
    resetComponentsState?: () => void;
  }

  const { t } = useI18n();
  const router = useRouter();
  const { currentHostId } = useCurrentHost();

  const activeTab = ref(CRONTAB_TYPE.System);
  const localListRef = ref<ListInstance | null>(null);
  const globalListRef = ref<ListInstance | null>(null);
  const systemListRef = ref<ListInstance | null>(null);

  const handleOpenConfigDirectory = () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error(t('common.host_id_required'));
      return;
    }
    router.push(createFileRoute('/etc/cron.d', { id: hostId }));
  };

  // 确保在tab切换时强制刷新列表
  const handleTabChange = async (key: string | number) => {
    // 使用 nextTick 确保DOM已更新
    await nextTick();

    if (key === CRONTAB_TYPE.Local && localListRef.value) {
      // 强制重置本地列表的状态
      localListRef.value.resetComponentsState?.();
    } else if (key === CRONTAB_TYPE.Global && globalListRef.value) {
      // 强制重置全局列表的状态
      globalListRef.value.resetComponentsState?.();
    } else if (key === CRONTAB_TYPE.System && systemListRef.value) {
      // 强制重置系统列表的状态
      systemListRef.value.resetComponentsState?.();
    }
  };

  onMounted(() => {
    // 确保初始加载时重置状态
    handleTabChange(activeTab.value);
  });
</script>
