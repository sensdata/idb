<template>
  <a-tabs v-model:active-key="activeTab" lazy-load @change="handleTabChange">
    <template #extra>
      <a-button
        v-if="activeTab === SERVICE_TYPE.System"
        size="small"
        @click="handleOpenConfigDirectory"
      >
        <template #icon>
          <icon-folder />
        </template>
        {{ $t('app.service.form.open_config_dir') }}
      </a-button>
    </template>
    <a-tab-pane
      :key="SERVICE_TYPE.System"
      :title="$t('app.service.enum.type.system')"
    >
      <list
        :key="`list-${SERVICE_TYPE.System}`"
        ref="systemListRef"
        :type="SERVICE_TYPE.System"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="SERVICE_TYPE.Local"
      :title="$t('app.service.enum.type.local')"
    >
      <list
        :key="`list-${SERVICE_TYPE.Local}`"
        ref="localListRef"
        :type="SERVICE_TYPE.Local"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="SERVICE_TYPE.Global"
      :title="$t('app.service.enum.type.global')"
    >
      <list
        :key="`list-${SERVICE_TYPE.Global}`"
        ref="globalListRef"
        :type="SERVICE_TYPE.Global"
      />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { Message } from '@arco-design/web-vue';
  import { IconFolder } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { SERVICE_TYPE } from '@/config/enum';
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

  const activeTab = ref(SERVICE_TYPE.System);
  const localListRef = ref<ListInstance | null>(null);
  const globalListRef = ref<ListInstance | null>(null);
  const systemListRef = ref<ListInstance | null>(null);

  const handleOpenConfigDirectory = () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error(t('common.host_id_required'));
      return;
    }
    router.push(createFileRoute('/etc/systemd/system', { id: hostId }));
  };

  // 确保在tab切换时刷新当前激活的列表
  const handleTabChange = async (key: string | number) => {
    // 使用 nextTick 确保DOM已更新
    await nextTick();

    // 切换tab时，刷新对应的列表以确保数据最新
    if (key === SERVICE_TYPE.System && systemListRef.value) {
      systemListRef.value.resetComponentsState?.();
    } else if (key === SERVICE_TYPE.Local && localListRef.value) {
      localListRef.value.resetComponentsState?.();
    } else if (key === SERVICE_TYPE.Global && globalListRef.value) {
      globalListRef.value.resetComponentsState?.();
    }
  };

  onMounted(() => {
    // 确保初始加载时重置状态
    handleTabChange(activeTab.value);
  });
</script>
