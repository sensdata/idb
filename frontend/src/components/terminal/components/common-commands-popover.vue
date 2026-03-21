<template>
  <a-popover
    v-model:popup-visible="visible"
    position="br"
    trigger="click"
    :content-style="{
      padding: 0,
      width: '400px',
      maxHeight: '500px',
    }"
  >
    <a-button type="text" size="small">
      <template #icon>
        <icon-code />
      </template>
      {{ $t('components.terminal.commands.title') }}
    </a-button>
    <template #content>
      <div class="commands-popover">
        <div class="popover-head">
          <div class="popover-title">
            {{ $t('components.terminal.commands.title') }}
          </div>
          <span class="arco-icon-hover popover-close" @click="handleClose">
            <icon-close />
          </span>
        </div>
        <div class="popover-body">
          <a-collapse
            :default-active-key="['0']"
            :bordered="false"
            expand-icon-position="right"
          >
            <a-collapse-item
              v-for="(category, catIndex) in commandCategories"
              :key="String(catIndex)"
              :header="category.category"
            >
              <div class="command-list">
                <div
                  v-for="(item, itemIndex) in category.items"
                  :key="itemIndex"
                  class="command-item"
                  @click="handleSelectCommand(item.command, item.autoExecute)"
                >
                  <span class="command-label">{{ item.label }}</span>
                  <code class="command-tag">{{ item.command }}</code>
                </div>
              </div>
            </a-collapse-item>
          </a-collapse>
        </div>
      </div>
    </template>
  </a-popover>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  interface CommandItem {
    label: string;
    command: string;
    autoExecute?: boolean;
  }

  interface CommandCategory {
    category: string;
    items: CommandItem[];
  }

  interface Emits {
    (e: 'select', command: string, autoExecute: boolean): void;
  }

  const emit = defineEmits<Emits>();
  const { t } = useI18n();

  const visible = ref(false);

  const commandCategories = computed<CommandCategory[]>(() => [
    {
      category: t('components.terminal.commands.categories.docker'),
      items: [
        {
          label: t('components.terminal.commands.items.dockerPs'),
          command:
            "docker ps --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'",
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.dockerComposePs'),
          command: 'docker compose ps',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.dockerComposeLogs'),
          command: 'docker compose logs -f',
        },
        {
          label: t('components.terminal.commands.items.dockerLogs'),
          command: 'docker logs -f <container>',
        },
        {
          label: t('components.terminal.commands.items.dockerInspect'),
          command: 'docker inspect <container>',
        },
      ],
    },
    {
      category: t('components.terminal.commands.categories.system'),
      items: [
        {
          label: t('components.terminal.commands.items.systemLoad'),
          command: 'uptime',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.systemMemory'),
          command: 'free -h',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.systemDisk'),
          command: 'df -h',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.systemFailed'),
          command: 'systemctl --failed',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.systemPorts'),
          command: 'ss -tulpen',
          autoExecute: true,
        },
      ],
    },
    {
      category: t('components.terminal.commands.categories.updates'),
      items: [
        {
          label: t('components.terminal.commands.items.updatesRefresh'),
          command: 'sudo apt update',
        },
        {
          label: t('components.terminal.commands.items.updatesList'),
          command: 'apt list --upgradable',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.updatesUpgrade'),
          command: 'sudo apt upgrade',
        },
        {
          label: t('components.terminal.commands.items.updatesFullUpgrade'),
          command: 'sudo apt full-upgrade',
        },
        {
          label: t('components.terminal.commands.items.updatesAutoStatus'),
          command: 'systemctl status unattended-upgrades --no-pager -l',
        },
        {
          label: t('components.terminal.commands.items.updatesAutoConfig'),
          command: 'sudo editor /etc/apt/apt.conf.d/20auto-upgrades',
        },
        {
          label: t('components.terminal.commands.items.updatesAutoEnable'),
          command: 'sudo dpkg-reconfigure -plow unattended-upgrades',
        },
      ],
    },
    {
      category: t('components.terminal.commands.categories.logs'),
      items: [
        {
          label: t('components.terminal.commands.items.logsJournalFollow'),
          command: 'sudo journalctl -f',
        },
        {
          label: t('components.terminal.commands.items.logsServiceDetail'),
          command: 'sudo journalctl -xe --no-pager',
        },
        {
          label: t('components.terminal.commands.items.logsKernel'),
          command: 'dmesg | tail -n 50',
          autoExecute: true,
        },
        {
          label: t('components.terminal.commands.items.logsCurlHealth'),
          command: 'curl -I http://127.0.0.1',
        },
        {
          label: t('components.terminal.commands.items.logsPingGateway'),
          command: 'ping -c 4 8.8.8.8',
        },
      ],
    },
  ]);

  function handleClose() {
    visible.value = false;
  }

  function handleSelectCommand(command: string, autoExecute?: boolean) {
    emit('select', command, autoExecute ?? false);
    visible.value = false;
  }
</script>

<style scoped>
  .commands-popover {
    display: flex;
    flex-direction: column;
    max-height: 28.125rem;
  }

  .popover-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.625rem 0.75rem;
    border-bottom: 0.0625rem solid var(--color-border-2);
  }

  .popover-title {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .popover-close {
    color: var(--color-text-3);
    cursor: pointer;
  }

  .popover-close:hover {
    color: var(--color-text-1);
  }

  .popover-body {
    flex: 1;
    padding: 0.25rem;
    overflow-y: auto;
  }

  .command-list {
    padding: 0.25rem 0;
  }

  .command-item {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0.625rem;
    cursor: pointer;
    border-radius: 0.25rem;
  }

  .command-item:hover {
    background: var(--color-fill-2);
  }

  .command-label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.8125rem;
    color: var(--color-text-2);
    white-space: nowrap;
  }

  .command-tag {
    flex-shrink: 0;
    max-width: 10rem;
    padding: 0.125rem 0.375rem;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: monospace;
    font-size: 0.75rem;
    color: var(--color-text-1);
    white-space: nowrap;
    background: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 0.1875rem;
  }

  :deep(.arco-collapse) {
    border: none;
  }

  :deep(.arco-collapse-item) {
    border: none;
  }

  :deep(.arco-collapse-item-header) {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    padding: 0.5rem 0.625rem;
    font-size: 0.8125rem;
    background: transparent;
    border: none;
  }

  :deep(.arco-collapse-item-header-left) {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  :deep(.arco-collapse-item-expand-icon) {
    flex-shrink: 0;
  }

  :deep(.arco-collapse-item-content) {
    padding: 0 !important;
    background: transparent;
  }

  :deep(.arco-collapse-item-content-box) {
    padding: 0 0.25rem;
  }
</style>
