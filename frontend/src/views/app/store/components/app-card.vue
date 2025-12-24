<template>
  <a-card hoverable>
    <div class="item-box flex gap-3 h-26">
      <a-avatar
        shape="square"
        class="item-avatar"
        :size="72"
        :style="{
          backgroundColor: getHexColorByChar(app.display_name),
        }"
      >
        {{ app.display_name.charAt(0) }}
      </a-avatar>
      <div class="item-main flex-1">
        <h3 class="mt-0 mb-3">
          {{ app.display_name }}
        </h3>
        <!-- 未安装应用显示描述和分类 -->
        <template v-if="app.status === 'uninstalled'">
          <div class="mb-4 text-sm text-gray-500 line-clamp-2">
            {{ app.description }}
          </div>
          <a-tag color="blue">{{ app.category }}</a-tag>
        </template>
        <!-- 已安装应用显示版本和安装时间 -->
        <template v-else>
          <a-tag bordered class="text-gray-600 mb-2">
            {{ $t('app.store.app.list.version') }}:
            {{ app.current_version }}
          </a-tag>
          <div class="text-gray-500 text-sm mb-2">
            {{ $t('app.store.app.list.install_at') }}:
            {{
              app.versions && app.versions[0] ? app.versions[0].created_at : ''
            }}
          </div>
        </template>
      </div>
      <div class="item-actions flex flex-col gap-3">
        <!-- 未安装应用显示安装按钮 -->
        <a-button
          v-if="app.status === 'uninstalled'"
          type="primary"
          shape="round"
          size="small"
          @click="$emit('install', app)"
        >
          {{ $t('app.store.app.list.install') }}
        </a-button>
        <!-- 已安装应用显示升级和卸载按钮 -->
        <template v-else>
          <a-button
            type="primary"
            shape="round"
            size="small"
            :disabled="!app.has_upgrade"
            @click="$emit('upgrade', app)"
          >
            {{ $t('app.store.app.list.upgrade') }}
          </a-button>
          <a-button
            type="primary"
            shape="round"
            status="danger"
            size="small"
            @click="$emit('uninstall', app)"
          >
            {{ $t('app.store.app.list.uninstall') }}
          </a-button>
        </template>
      </div>
    </div>
  </a-card>
</template>

<script setup lang="ts">
  import { AppSimpleEntity } from '@/entity/App';
  import { getHexColorByChar } from '@/helper/utils';

  defineProps<{
    app: AppSimpleEntity;
  }>();

  defineEmits<{
    install: [app: AppSimpleEntity];
    upgrade: [app: AppSimpleEntity];
    uninstall: [app: AppSimpleEntity];
  }>();
</script>
