<template>
  <div class="ssh-content">
    <div class="form-item">
      <div class="label">{{ $t('app.ssh.port.label') }}</div>
      <div class="content with-input-group">
        <div class="input-group">
          <a-input
            v-model="port"
            placeholder="22"
            disabled
            class="short-input"
          />
          <div class="actions">
            <a-button
              type="text"
              class="setting-btn"
              @click="openSettingModal('port')"
            >
              <icon-settings />
              <span>{{ $t('app.ssh.btn.setting') }}</span>
            </a-button>
          </div>
        </div>
        <div class="description">{{ $t('app.ssh.port.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.listen.label') }}</div>
      <div class="content with-input-group">
        <div class="input-group">
          <a-input
            v-model="listenAddress"
            placeholder="0.0.0.0"
            disabled
            class="short-input"
          />
          <div class="actions">
            <a-button
              type="text"
              class="setting-btn"
              @click="openSettingModal('listen')"
            >
              <icon-settings />
              <span>{{ $t('app.ssh.btn.setting') }}</span>
            </a-button>
          </div>
        </div>
        <div class="description">{{ $t('app.ssh.listen.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.root.label') }}</div>
      <div class="content with-input-group">
        <div class="input-group">
          <a-input
            v-model="rootUser"
            placeholder="允许 SSH 登录"
            disabled
            class="short-input"
          />
          <div class="actions">
            <a-button
              type="text"
              class="setting-btn"
              @click="openSettingModal('root')"
            >
              <icon-settings />
              <span>{{ $t('app.ssh.btn.setting') }}</span>
            </a-button>
          </div>
        </div>
        <div class="description">{{ $t('app.ssh.root.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.password.label') }}</div>
      <div class="content">
        <a-switch v-model="passwordAuth" />
        <div class="description">{{ $t('app.ssh.password.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.key.label') }}</div>
      <div class="content">
        <a-switch v-model="keyAuth" />
        <div class="description">{{ $t('app.ssh.key.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.passwordInfo.label') }}</div>
      <div class="content">
        <div class="password-info-placeholder"></div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.reverse.label') }}</div>
      <div class="content">
        <a-switch v-model="reverseLookup" />
        <div class="description">{{ $t('app.ssh.reverse.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">{{ $t('app.ssh.autostart.label') }}</div>
      <div class="content">
        <a-switch v-model="autoStart" />
      </div>
    </div>

    <!-- 端口设置弹窗 -->
    <a-modal
      v-model:visible="portModalVisible"
      :title="$t('app.ssh.portModal.title')"
      @ok="savePortSetting"
      @cancel="portModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <div class="modal-label">{{ $t('app.ssh.port.label') }}</div>
          <div class="modal-input-wrapper">
            <a-input v-model="portForm.port" placeholder="22" />
            <div class="modal-field-description">{{
              $t('app.ssh.port.description')
            }}</div>
          </div>
        </div>
      </div>
    </a-modal>

    <!-- 监听地址设置弹窗 -->
    <a-modal
      v-model:visible="listenModalVisible"
      :title="$t('app.ssh.listenModal.title')"
      @ok="saveListenSetting"
      @cancel="listenModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <div class="modal-label">{{ $t('app.ssh.listen.label') }}</div>
          <div class="modal-input-wrapper">
            <a-input v-model="listenForm.address" placeholder="0.0.0.0" />
            <div class="modal-field-description">{{
              $t('app.ssh.listen.description')
            }}</div>
          </div>
        </div>
      </div>
    </a-modal>

    <!-- Root用户设置弹窗 -->
    <a-modal
      v-model:visible="rootModalVisible"
      :title="$t('app.ssh.rootModal.title')"
      @ok="saveRootSetting"
      @cancel="rootModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <div class="modal-label">{{ $t('app.ssh.root.label') }}</div>
          <div class="modal-input-wrapper">
            <a-radio-group v-model="rootForm.enabled">
              <a-radio :value="true">{{
                $t('app.ssh.rootModal.allow')
              }}</a-radio>
              <a-radio :value="false">{{
                $t('app.ssh.rootModal.deny')
              }}</a-radio>
            </a-radio-group>
            <div class="modal-field-description">{{
              $t('app.ssh.root.description')
            }}</div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconSettings } from '@arco-design/web-vue/es/icon';
  import { Message } from '@arco-design/web-vue';

  const { t } = useI18n();

  // 表单的默认值
  const port = ref('22');
  const listenAddress = ref('0.0.0.0');
  const rootUser = ref('允许 SSH 登录');
  const passwordAuth = ref(true);
  const keyAuth = ref(true);
  const reverseLookup = ref(true);
  const autoStart = ref(true);

  // 弹窗显示状态
  const portModalVisible = ref(false);
  const listenModalVisible = ref(false);
  const rootModalVisible = ref(false);

  // 弹窗表单数据
  const portForm = ref({
    port: '22',
  });

  const listenForm = ref({
    address: '0.0.0.0',
  });

  const rootForm = ref({
    enabled: true,
  });

  // 初始化数据
  onMounted(async () => {
    try {
      // 在这里可以添加实际的API调用来获取SSH服务配置
      // 例如: const response = await getSSHConfig();
      // 然后用获取的数据更新本地状态
    } catch (error) {
      // 错误处理
    }
  });

  // 打开设置弹窗函数
  const openSettingModal = (type: 'port' | 'listen' | 'root') => {
    if (type === 'port') {
      portForm.value.port = port.value;
      portModalVisible.value = true;
    } else if (type === 'listen') {
      listenForm.value.address = listenAddress.value;
      listenModalVisible.value = true;
    } else if (type === 'root') {
      rootForm.value.enabled = rootUser.value === '允许 SSH 登录';
      rootModalVisible.value = true;
    }
  };

  // 保存各项设置的函数
  const savePortSetting = () => {
    port.value = portForm.value.port;
    portModalVisible.value = false;
    Message.success(t('app.ssh.portModal.saveSuccess'));
  };

  const saveListenSetting = () => {
    listenAddress.value = listenForm.value.address;
    listenModalVisible.value = false;
    Message.success(t('app.ssh.listenModal.saveSuccess'));
  };

  const saveRootSetting = () => {
    rootUser.value = rootForm.value.enabled ? '允许 SSH 登录' : '禁止 SSH 登录';
    rootModalVisible.value = false;
    Message.success(t('app.ssh.rootModal.saveSuccess'));
  };
</script>

<style scoped lang="less">
  .ssh-content {
    margin-top: 8px;
  }

  .form-item {
    display: flex;
    align-items: flex-start;
    margin-bottom: 24px;

    .label {
      flex-shrink: 0;
      width: 100px;
      margin-right: 16px;
      color: var(--color-text-1);
      font-weight: 500;
      line-height: 32px;
      text-align: right;
    }

    .content {
      display: flex;
      flex: 1;

      &.with-input-group {
        flex-direction: column;

        .input-group {
          display: flex;
          align-items: center;
          margin-bottom: 4px;

          .short-input {
            width: 200px;
          }
        }

        .description {
          color: var(--color-text-3);
          font-size: 12px;
        }
      }

      &:not(.with-input-group) {
        align-items: center;

        .description {
          margin-left: 8px;
          color: var(--color-text-3);
          font-size: 12px;
        }
      }
    }

    .actions {
      display: flex;
      flex-shrink: 0;
      align-items: center;
      justify-content: center;
      height: 32px;
      margin-left: 8px;
      text-align: center;

      .setting-btn {
        padding: 0 8px;
        color: #8250df; /* 紫色按钮文字 */
        background: none;
        border: none;
        box-shadow: none;

        &:hover {
          color: #9e77e3; /* 悬停时的浅紫色 */
          background: none;
        }
      }
    }

    .password-info-placeholder {
      width: 200px;
      height: 32px;
      background-color: var(--color-fill-2);
      border-radius: 4px;
    }
  }

  .modal-form-wrapper {
    padding: 0 20px;
  }

  .modal-form-item {
    display: flex;
    margin-bottom: 20px;
  }

  .modal-label {
    width: 80px;
    margin-right: 20px;
    color: var(--color-text-1);
    font-weight: 500;
    line-height: 32px;
    text-align: right;
  }

  .modal-input-wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .modal-field-description {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
