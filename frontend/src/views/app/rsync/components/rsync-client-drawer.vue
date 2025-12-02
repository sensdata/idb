<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="
      isEdit
        ? $t('app.rsync.client.action.edit')
        : $t('app.rsync.client.action.create')
    "
    :width="600"
    unmount-on-close
    @cancel="handleCancel"
  >
    <template #footer>
      <div class="drawer-footer">
        <a-button @click="handleCancel">
          {{ $t('common.cancel') }}
        </a-button>
        <a-button type="primary" :loading="loading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </a-button>
      </div>
    </template>

    <a-form ref="formRef" :model="form" :rules="rules" layout="vertical">
      <a-form-item field="name" :label="$t('app.rsync.client.form.name')">
        <a-input
          v-model="form.name"
          :placeholder="$t('app.rsync.client.form.placeholder.name')"
        />
      </a-form-item>

      <a-form-item
        field="direction"
        :label="$t('app.rsync.client.form.direction')"
      >
        <a-radio-group v-model="form.direction">
          <a-radio value="local_to_remote">
            {{ $t('app.rsync.client.direction.localToRemote') }}
          </a-radio>
          <a-radio value="remote_to_local">
            {{ $t('app.rsync.client.direction.remoteToLocal') }}
          </a-radio>
        </a-radio-group>
      </a-form-item>

      <a-form-item
        field="local_path"
        :label="$t('app.rsync.client.form.localPath')"
      >
        <file-selector
          v-model="form.local_path"
          :placeholder="$t('app.rsync.client.form.placeholder.localPath')"
          type="all"
        />
      </a-form-item>

      <a-form-item
        field="remote_type"
        :label="$t('app.rsync.client.form.remoteType')"
      >
        <a-radio-group v-model="form.remote_type">
          <a-radio value="rsync">
            {{ $t('app.rsync.client.remoteType.rsync') }}
          </a-radio>
          <a-radio value="ssh">
            {{ $t('app.rsync.client.remoteType.ssh') }}
          </a-radio>
        </a-radio-group>
      </a-form-item>

      <a-row :gutter="16">
        <a-col :span="16">
          <a-form-item
            field="remote_host"
            :label="$t('app.rsync.client.form.remoteHost')"
          >
            <a-input
              v-model="form.remote_host"
              :placeholder="$t('app.rsync.client.form.placeholder.remoteHost')"
            />
          </a-form-item>
        </a-col>
        <a-col :span="8">
          <a-form-item
            field="remote_port"
            :label="$t('app.rsync.client.form.remotePort')"
          >
            <a-input-number
              v-model="form.remote_port"
              :placeholder="$t('app.rsync.client.form.placeholder.remotePort')"
              :min="1"
              :max="65535"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-form-item
        field="auth_mode"
        :label="$t('app.rsync.client.form.authMode')"
      >
        <a-radio-group v-model="form.auth_mode">
          <a-radio value="password">
            {{ $t('app.rsync.client.authMode.password') }}
          </a-radio>
          <a-radio value="anonymous">
            {{ $t('app.rsync.client.authMode.anonymous') }}
          </a-radio>
          <a-radio v-if="form.remote_type === 'ssh'" value="private_key">
            {{ $t('app.rsync.client.authMode.privateKey') }}
          </a-radio>
        </a-radio-group>
      </a-form-item>

      <a-form-item
        v-if="form.auth_mode !== 'anonymous'"
        field="username"
        :label="$t('app.rsync.client.form.username')"
      >
        <a-input
          v-model="form.username"
          :placeholder="$t('app.rsync.client.form.placeholder.username')"
        />
      </a-form-item>

      <a-form-item
        v-if="form.auth_mode === 'password'"
        field="password"
        :label="$t('app.rsync.client.form.password')"
      >
        <a-input-password
          v-model="form.password"
          :placeholder="$t('app.rsync.client.form.placeholder.password')"
        />
      </a-form-item>

      <a-form-item
        v-if="form.auth_mode === 'private_key'"
        field="ssh_private_key"
        :label="$t('app.rsync.client.form.sshPrivateKey')"
      >
        <a-textarea
          v-model="form.ssh_private_key"
          :placeholder="$t('app.rsync.client.form.placeholder.sshPrivateKey')"
          :auto-size="{ minRows: 4, maxRows: 8 }"
        />
      </a-form-item>

      <a-form-item
        v-if="form.remote_type === 'rsync'"
        field="module"
        :label="$t('app.rsync.client.form.module')"
      >
        <a-input
          v-model="form.module"
          :placeholder="$t('app.rsync.client.form.placeholder.module')"
        />
      </a-form-item>

      <a-form-item
        field="remote_path"
        :label="
          form.remote_type === 'rsync'
            ? $t('app.rsync.client.form.remotePathInModule')
            : $t('app.rsync.client.form.remotePath')
        "
      >
        <a-input
          v-model="form.remote_path"
          :placeholder="
            form.remote_type === 'rsync'
              ? $t('app.rsync.client.form.placeholder.remotePathInModule')
              : $t('app.rsync.client.form.placeholder.remotePath')
          "
        />
      </a-form-item>

      <a-form-item field="enqueue" :label="$t('app.rsync.client.form.enqueue')">
        <a-switch v-model="form.enqueue" />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import { createRsyncClientTaskApi } from '@/api/database';
  import {
    RsyncClientTask,
    RsyncClientCreateTaskRequest,
  } from '@/entity/Database';
  import FileSelector from '@/components/file/file-selector/index.vue';

  interface Props {
    visible: boolean;
    editData?: RsyncClientTask | null;
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    editData: null,
  });

  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'ok'): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const loading = ref(false);

  const isEdit = computed(() => !!props.editData);

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  const getDefaultForm = (): RsyncClientCreateTaskRequest => ({
    name: '',
    direction: 'local_to_remote',
    local_path: '',
    remote_type: 'rsync',
    remote_host: '',
    remote_port: 873,
    username: '',
    auth_mode: 'password',
    password: '',
    ssh_private_key: '',
    remote_path: '',
    module: '',
    enqueue: true,
  });

  const form = ref<RsyncClientCreateTaskRequest>(getDefaultForm());

  const rules = {
    name: [{ required: true, message: t('common.form.required') }],
    direction: [{ required: true, message: t('common.form.required') }],
    local_path: [{ required: true, message: t('common.form.required') }],
    remote_type: [{ required: true, message: t('common.form.required') }],
    remote_host: [{ required: true, message: t('common.form.required') }],
    remote_port: [{ required: true, message: t('common.form.required') }],
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        if (props.editData) {
          form.value = {
            name: props.editData.name,
            direction: props.editData.direction,
            local_path: props.editData.local_path,
            remote_type: props.editData.remote_type,
            remote_host: props.editData.remote_host,
            remote_port: props.editData.remote_port,
            username: props.editData.username,
            auth_mode: props.editData.auth_mode as
              | 'password'
              | 'anonymous'
              | 'private_key',
            password: props.editData.password || '',
            ssh_private_key: props.editData.ssh_private_key || '',
            remote_path: props.editData.remote_path,
            module: props.editData.module || '',
            enqueue: true,
          };
        } else {
          form.value = getDefaultForm();
        }
      }
    }
  );

  watch(
    () => form.value.remote_type,
    (type) => {
      form.value.remote_port = type === 'rsync' ? 873 : 22;
      if (type === 'rsync' && form.value.auth_mode === 'private_key') {
        form.value.auth_mode = 'password';
      }
    }
  );

  const handleCancel = () => {
    drawerVisible.value = false;
  };

  const handleSubmit = async () => {
    const valid = await formRef.value?.validate();
    if (valid) return;

    loading.value = true;
    try {
      await createRsyncClientTaskApi(form.value);
      Message.success(
        isEdit.value
          ? t('app.rsync.client.message.updateSuccess')
          : t('app.rsync.client.message.createSuccess')
      );
      drawerVisible.value = false;
      emit('ok');
    } catch (error) {
      Message.error(
        isEdit.value
          ? t('app.rsync.client.message.updateFailed')
          : t('app.rsync.client.message.createFailed')
      );
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped lang="less">
  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.86rem;
  }
</style>
