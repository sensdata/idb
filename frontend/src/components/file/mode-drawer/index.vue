<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.modeDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item field="path" :label="$t('components.file.modeDrawer.path')">
        <span>{{ formState.path }}</span>
      </a-form-item>
      <a-form-item
        field="owner_access"
        :label="$t('components.file.modeDrawer.owner_access')"
      >
        <a-checkbox-group
          v-model="formState.owner_access"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item
        field="group_access"
        :label="$t('components.file.modeDrawer.group_access')"
      >
        <a-checkbox-group
          v-model="formState.group_access"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item
        field="other_access"
        :label="$t('components.file.modeDrawer.other_access')"
      >
        <a-checkbox-group
          v-model="formState.other_access"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item field="mode" :label="$t('components.file.modeDrawer.mode')">
        <a-input v-model="formState.mode" class="w-60" :max-length="4" />
      </a-form-item>
      <a-form-item field="user" :label="$t('components.file.modeDrawer.user')">
        <a-input v-model="formState.user" class="w-60" />
      </a-form-item>
      <a-form-item
        field="group"
        :label="$t('components.file.modeDrawer.group')"
      >
        <a-input v-model="formState.group" class="w-60" />
      </a-form-item>
      <a-form-item field="sub" label=" ">
        <a-checkbox v-model="formState.sub">
          {{ $t('components.file.modeDrawer.sub') }}
        </a-checkbox>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/hooks/loading';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { batchUpdateFileRoleApi } from '@/api/file';
  import { Message } from '@arco-design/web-vue';

  const { t } = useI18n();

  const emit = defineEmits(['ok']);

  const formRef = ref();
  const formState = reactive({
    path: '',
    mode: '0755',
    user: '',
    group: '',
    owner_access: ['4', '2', '1'],
    group_access: ['4', '1'],
    other_access: ['4', '1'],
    sub: true,
  });

  const accessOptions = [
    { label: t('components.file.modeDrawer.read'), value: '4' },
    { label: t('components.file.modeDrawer.write'), value: '2' },
    { label: t('components.file.modeDrawer.execute'), value: '1' },
  ];

  const rules = {
    mode: {
      required: true,
      validator: (value: string, cb: any) => {
        if (!/^0?[0-7]{3}$/.test(value)) {
          cb(t('components.file.modeDrawer.mode_error'));
        }
        cb();
      },
    },
    user: {
      required: true,
      message: t('components.file.modeDrawer.user_required'),
    },
    group: {
      required: true,
      message: t('components.file.modeDrawer.group_required'),
    },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  const showLoading = () => {
    setLoading(true);
  };

  const hideLoading = () => {
    setLoading(false);
  };

  const calculateMode = (access: string[]) => {
    return access.reduce((sum, per) => sum + Number(per), 0);
  };

  const calculateAccess = (digit: string) => {
    const arr: string[] = [];
    const n = parseInt(digit, 10); // 将字符转换为数字
    if (n & 4) {
      arr.push('4');
    }
    if (n & 2) {
      arr.push('2');
    }
    if (n & 1) {
      arr.push('1');
    }
    return arr;
  };

  let isUpdatingMode = false;
  let isUpdatingAccess = false;
  watch(
    () => [
      formState.owner_access,
      formState.group_access,
      formState.other_access,
    ],
    ([newOwner, newGroup, newOthers]) => {
      if (isUpdatingAccess) {
        return;
      }
      isUpdatingMode = true;
      const owner = calculateMode(newOwner);
      const group = calculateMode(newGroup);
      const others = calculateMode(newOthers);
      formState.mode = `0${owner}${group}${others}`;
      isUpdatingMode = false;
    },
    { deep: true }
  );

  watch(
    () => formState.mode,
    (newMode) => {
      if (isUpdatingMode) {
        return;
      }
      if (!/^0[0-7]{3}$/.test(newMode)) {
        return;
      }
      isUpdatingAccess = true;
      const [owner, group, others] = newMode.slice(1).split('');
      formState.owner_access = calculateAccess(owner);
      formState.group_access = calculateAccess(group);
      formState.other_access = calculateAccess(others);
      isUpdatingAccess = false;
    }
  );

  const setData = (data: FileInfoEntity) => {
    let mode = data.mode || '0755';
    mode = mode.padStart(4, '0');

    formState.path = data.path;
    formState.mode = mode;
    formState.user = data.user;
    formState.group = data.group;
    formState.owner_access = calculateAccess(mode.charAt(1));
    formState.group_access = calculateAccess(mode.charAt(2));
    formState.other_access = calculateAccess(mode.charAt(3));
  };

  const getData = () => {
    return {
      sources: [formState.path],
      mode: formState.mode,
      user: formState.user,
      group: formState.group,
      sub: formState.sub,
    };
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      await batchUpdateFileRoleApi({
        ...getData(),
        mode: +formState.mode,
      });
      Message.success(t('components.file.modeDrawer.message.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
