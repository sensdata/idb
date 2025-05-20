<template>
  <div class="container">
    <ssh-status class="ssh-status-container" />
    <a-tabs default-active-key="config">
      <a-tab-pane key="config" :title="$t('app.ssh.tabs.config')">
        <config-tab />
      </a-tab-pane>
      <a-tab-pane key="password" :title="$t('app.ssh.tabs.password')">
        <password-tab />
      </a-tab-pane>
      <a-tab-pane key="authkey" :title="$t('app.ssh.tabs.authkey')">
        <auth-key-tab />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script lang="ts" setup>
  import SshStatus from './components/ssh-status/index.vue';
  import PasswordTab from './components/password-tab/index.vue';
  import AuthKeyTab from './components/auth-key-tab/index.vue';
  import ConfigTab from './components/config-tab/index.vue';
</script>

<style scoped lang="less">
  .container {
    padding: 0 16px;
    background-color: var(--color-bg-2);
    border-radius: 4px;
    position: relative;
  }

  .ssh-status-container {
    margin-bottom: 16px;
    padding: 12px 0;
  }
</style>
