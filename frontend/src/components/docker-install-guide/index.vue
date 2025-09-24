<template>
  <div class="docker-install-guide">
    <!-- Docker 状态检测 -->
    <a-card
      v-if="showStatus && dockerStore.isNotInstalled"
      class="mb-4"
      :loading="loading"
    >
      <template #title>
        <div class="flex items-center gap-2">
          <IconComputer />
          <span>{{ $t('docker.install.guide.title') }}</span>
        </div>
      </template>

      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <a-tag :color="statusColor" :bordered="false" class="px-3 py-1">
            <template #icon>
              <IconCloseCircle />
            </template>
            {{ statusText }}
          </a-tag>

          <span class="text-gray-600">
            {{ statusDescription }}
          </span>
        </div>

        <div class="flex items-center gap-2">
          <a-button size="small" :loading="loading" @click="checkDockerStatus">
            <template #icon>
              <IconRefresh />
            </template>
            {{ $t('docker.install.guide.refresh') }}
          </a-button>

          <a-button
            v-if="dockerStore.isNotInstalled"
            type="primary"
            size="small"
            :loading="installing"
            @click="handleInstallDocker"
          >
            <template #icon>
              <IconDownload />
            </template>
            {{ $t('docker.install.guide.install') }}
          </a-button>
        </div>
      </div>
    </a-card>

    <!-- 安装进度弹窗 -->
    <a-modal
      v-model:visible="installModalVisible"
      :title="$t('docker.install.guide.installing.title')"
      :closable="false"
      :mask-closable="false"
      :footer="false"
      width="600px"
    >
      <div class="install-progress">
        <div class="flex items-center gap-3 mb-4">
          <a-spin :loading="installing" />
          <span class="text-lg">{{ installStatusText }}</span>
        </div>

        <a-progress
          v-if="installing"
          :percent="0"
          :status="installError ? 'danger' : 'normal'"
          :show-text="false"
          animation
        />
        <div v-else-if="installSuccess" class="flex items-center gap-2">
          <a-progress
            :percent="100"
            status="success"
            :show-text="false"
            class="flex-1"
          />
          <span class="text-sm text-green-600 font-medium">100%</span>
        </div>
        <div v-else-if="installError" class="flex items-center gap-2">
          <a-progress
            :percent="100"
            status="danger"
            :show-text="false"
            class="flex-1"
          />
          <span class="text-sm text-red-600 font-medium">100%</span>
        </div>

        <div v-if="installLogs.length > 0" class="mt-4">
          <div class="bg-gray-100 p-3 rounded max-h-60 overflow-y-auto">
            <div
              v-for="(log, index) in installLogs"
              :key="index"
              class="text-sm font-mono mb-1"
              :class="{
                'text-red-600': log.includes('error') || log.includes('failed'),
                'text-green-600':
                  log.includes('success') || log.includes('完成'),
                'text-gray-700': true,
              }"
            >
              {{ log }}
            </div>
          </div>
        </div>

        <div v-if="installSuccess" class="mt-4 flex justify-end">
          <a-button type="primary" @click="handleInstallComplete">
            {{ $t('docker.install.guide.install.complete') }}
          </a-button>
        </div>

        <div v-if="installError" class="mt-4 flex justify-end">
          <a-button @click="handleInstallError">
            {{ $t('common.close') }}
          </a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    IconComputer,
    IconCloseCircle,
    IconRefresh,
    IconDownload,
  } from '@arco-design/web-vue/es/icon';
  import { dockerInstallApi } from '@/api/docker';
  import useDockerStatusStore from '@/store/modules/docker';
  import { useRoute } from 'vue-router';

  interface Props {
    showStatus?: boolean;
  }

  withDefaults(defineProps<Props>(), {
    showStatus: true,
  });

  const emit = defineEmits<{
    statusChange: [status: string];
    installComplete: [];
  }>();

  const { t } = useI18n();
  const dockerStore = useDockerStatusStore();
  const route = useRoute();
  watch(
    () => route.query.id,
    (id) => {
      dockerStore.setCurrentHost(id ? String(id) : null);
    },
    { immediate: true }
  );

  // 状态管理
  const loading = ref(false);
  const installing = ref(false);
  const dockerStatus = ref<string>('');
  const installModalVisible = ref(false);
  const installStatusText = ref('');
  const installLogs = ref<string[]>([]);
  const installError = ref('');
  const installSuccess = ref(false);

  // 计算属性
  const statusColor = computed(() => {
    switch (dockerStore.currentStatus) {
      case 'installed':
        return 'rgb(var(--success-6))';
      case 'not installed':
        return 'rgb(var(--danger-6))';
      default:
        return 'rgb(var(--color-text-4))';
    }
  });

  const statusText = computed(() => {
    switch (dockerStore.currentStatus) {
      case 'installed':
        return t('docker.install.guide.status.installed');
      case 'not installed':
        return t('docker.install.guide.status.not_installed');
      default:
        return t('docker.install.guide.status.checking');
    }
  });

  const statusDescription = computed(() => {
    switch (dockerStore.currentStatus) {
      case 'installed':
        return t('docker.install.guide.status.installed.desc');
      case 'not installed':
        return t('docker.install.guide.status.not_installed.desc');
      default:
        return t('docker.install.guide.status.checking.desc');
    }
  });

  // 检查 Docker 安装状态
  const checkDockerStatus = async () => {
    loading.value = true;
    try {
      await dockerStore.refresh();
      dockerStatus.value = dockerStore.currentStatus;
      emit('statusChange', dockerStore.currentStatus);
    } catch (error: any) {
      console.error('检查 Docker 状态失败:', error);
      Message.error(t('docker.install.guide.check.failed'));
    } finally {
      loading.value = false;
    }
  };

  // 安装 Docker
  const handleInstallDocker = async () => {
    installModalVisible.value = true;
    installing.value = true;
    installStatusText.value = t('docker.install.guide.installing.preparing');
    installLogs.value = [];
    installError.value = '';
    installSuccess.value = false;

    try {
      // 显示开始安装的日志
      installLogs.value.push(
        `[${new Date().toLocaleTimeString()}] ${t(
          'docker.install.guide.installing.preparing'
        )}`
      );

      // 设置为正在安装状态
      installStatusText.value = t('docker.install.guide.installing.installing');
      installLogs.value.push(
        `[${new Date().toLocaleTimeString()}] ${t(
          'docker.install.guide.installing.installing'
        )}`
      );

      // 调用安装 API - 这是真实的安装过程
      await dockerInstallApi();

      // 安装成功
      installSuccess.value = true;
      installStatusText.value = t('docker.install.guide.installing.success');
      installLogs.value.push(
        `[${new Date().toLocaleTimeString()}] Docker 安装成功！`
      );

      // 重新检查状态
      await checkDockerStatus();
    } catch (error: any) {
      console.error('Docker 安装失败:', error);
      installError.value =
        error.message || t('docker.install.guide.install.failed');
      installStatusText.value = t('docker.install.guide.install.failed');
      installLogs.value.push(
        `[${new Date().toLocaleTimeString()}] 安装失败: ${error.message}`
      );
    } finally {
      installing.value = false;
    }
  };

  // 安装完成处理
  const handleInstallComplete = () => {
    installModalVisible.value = false;
    emit('installComplete');
    Message.success(t('docker.install.guide.install.success'));
  };

  // 安装失败处理
  const handleInstallError = () => {
    installModalVisible.value = false;
    // 重置安装状态
    installError.value = '';
    installLogs.value = [];
    installSuccess.value = false;
  };

  // 监听状态变化
  watch(
    () => dockerStore.currentStatus,
    (newStatus) => {
      emit('statusChange', newStatus);
    }
  );

  // 组件挂载时不自动检查，只有在明确需要时才检查

  // 暴露方法给父组件
  defineExpose({
    checkDockerStatus,
    dockerStatus: computed(() => dockerStore.currentStatus),
  });
</script>

<style scoped>
  .docker-install-guide {
    width: 100%;
  }

  .install-progress {
    padding: 16px 0;
  }

  .text-sm {
    font-size: 12px;
    line-height: 1.4;
  }

  .font-mono {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
  }

  .max-h-60 {
    max-height: 15rem;
  }

  .overflow-y-auto {
    overflow-y: auto;
  }
</style>
