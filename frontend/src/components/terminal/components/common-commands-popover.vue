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
                  @click="
                    handleSelectCommand(item.command, category.autoExecute)
                  "
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
  import { ref } from 'vue';

  interface CommandItem {
    label: string;
    command: string;
  }

  interface CommandCategory {
    category: string;
    autoExecute?: boolean;
    items: CommandItem[];
  }

  interface Emits {
    (e: 'select', command: string, autoExecute: boolean): void;
  }

  const emit = defineEmits<Emits>();

  const visible = ref(false);

  const commandCategories: CommandCategory[] = [
    {
      category: '系统信息',
      autoExecute: true,
      items: [
        { label: '查看 CPU/内存/进程 (top)', command: 'top' },
        { label: '查看内存使用 (free -h)', command: 'free -h' },
        { label: '查看磁盘使用 (df -h)', command: 'df -h' },
        { label: '查看目录大小 (du -sh *)', command: 'du -sh *' },
        { label: '查看系统版本 (uname -a)', command: 'uname -a' },
        { label: '查看系统运行时长 (uptime)', command: 'uptime' },
        { label: '查看主机名与系统版本 (hostnamectl)', command: 'hostnamectl' },
      ],
    },
    {
      category: '网络诊断',
      items: [
        { label: 'Ping 测试网络 (ping)', command: 'ping -c 4 8.8.8.8' },
        {
          label: '测试网页是否可访问 (curl -I)',
          command: 'curl -I https://example.com',
        },
        { label: '查看端口占用 (ss -tulnp)', command: 'ss -tulnp' },
        { label: '查看指定端口 (ss | grep)', command: 'ss -anp | grep :80' },
        { label: '查看 IP 信息 (ip a)', command: 'ip a' },
        { label: '查看路由 (ip r)', command: 'ip r' },
        { label: '查询 DNS (nslookup)', command: 'nslookup example.com' },
      ],
    },
    {
      category: '文件与目录操作',
      items: [
        { label: '查看文件(权限) (ls -al)', command: 'ls -al' },
        { label: '打印当前路径 (pwd)', command: 'pwd' },
        { label: '创建目录 (mkdir -p)', command: 'mkdir -p new_folder' },
        { label: '复制文件 (cp)', command: 'cp source.txt dest.txt' },
        { label: '移动/重命名文件 (mv)', command: 'mv oldname newname' },
        { label: '删除文件/目录 (rm -rf)', command: 'rm -rf /path/to/remove' },
        { label: '查看文件内容 (cat)', command: 'cat file.txt' },
        { label: '实时查看日志 (tail -f)', command: 'tail -f /var/log/syslog' },
        { label: '查找文件 (find)', command: 'find / -name filename' },
      ],
    },
    {
      category: '权限/用户管理',
      items: [
        { label: '修改权限 (chmod)', command: 'chmod 755 file' },
        { label: '修改属主 (chown)', command: 'chown user:group file' },
        { label: '添加用户 (useradd)', command: 'useradd newuser' },
        { label: '修改密码 (passwd)', command: 'passwd username' },
      ],
    },
    {
      category: '服务与进程管理',
      items: [
        { label: '查看所有进程 (ps -aux)', command: 'ps -aux' },
        { label: '查找进程 (ps | grep)', command: 'ps -ef | grep nginx' },
        {
          label: '服务状态 (systemctl status)',
          command: 'systemctl status nginx',
        },
        {
          label: '启动服务 (systemctl start)',
          command: 'systemctl start nginx',
        },
        { label: '停止服务 (systemctl stop)', command: 'systemctl stop nginx' },
        {
          label: '重启服务 (systemctl restart)',
          command: 'systemctl restart nginx',
        },
        {
          label: '设为开机自启 (systemctl enable)',
          command: 'systemctl enable nginx',
        },
      ],
    },
    {
      category: '压缩与传输',
      items: [
        {
          label: '压缩目录 (tar.gz)',
          command: 'tar -czf archive.tar.gz directory/',
        },
        { label: '解压 tar.gz', command: 'tar -xzf archive.tar.gz' },
        {
          label: '远程复制文件 (scp)',
          command: 'scp file.txt user@host:/path/',
        },
      ],
    },
    {
      category: '日志与排查',
      items: [
        {
          label: '查看服务日志 (journalctl -u)',
          command: 'journalctl -u nginx -f',
        },
        { label: '查看内核日志 (dmesg | tail)', command: 'dmesg | tail' },
        { label: '系统 IO 状况 (iotop)', command: 'iotop' },
      ],
    },
  ];

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
