<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('app.file.modeDrawer.title')"
    unmountOnClose
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form :model="formState" :rules="rules">
      <a-form-item field="path" :label="$t('app.file.modeDrawer.path')">
        <span>{{ formState.path }}</span>
      </a-form-item>
      <a-form-item
        field="ownerAccess"
        :label="$t('app.file.modeDrawer.ownerAccess')"
      >
        <a-checkbox-group
          v-model="formState.ownerAccess"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item
        field="groupAccess"
        :label="$t('app.file.modeDrawer.groupAccess')"
      >
        <a-checkbox-group
          v-model="formState.groupAccess"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item
        field="otherAccess"
        :label="$t('app.file.modeDrawer.otherAccess')"
      >
        <a-checkbox-group
          v-model="formState.otherAccess"
          :options="accessOptions"
        />
      </a-form-item>
      <a-form-item field="mode" :label="$t('app.file.modeDrawer.mode')">
        <a-input v-model="formState.mode" class="w-60" :max-length="4" />
      </a-form-item>
      <a-form-item field="user" :label="$t('app.file.modeDrawer.user')">
        <a-input v-model="formState.user" class="w-60" />
      </a-form-item>
      <a-form-item field="group" :label="$t('app.file.modeDrawer.group')">
        <a-input v-model="formState.group" class="w-60" />
      </a-form-item>
      <a-form-item field="sub" label=" ">
        <a-checkbox v-model="formState.sub">
          {{ $t('app.file.modeDrawer.sub') }}
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

  const { t } = useI18n();

  const formState = reactive({
    path: '',
    mode: '0755',
    user: '',
    group: '',
    ownerAccess: ['4', '2', '1'],
    groupAccess: ['4', '1'],
    otherAccess: ['4', '1'],
    sub: true,
  });

  const accessOptions = [
    { label: t('app.file.modeDrawer.read'), value: '4' },
    { label: t('app.file.modeDrawer.write'), value: '2' },
    { label: t('app.file.modeDrawer.execute'), value: '1' },
  ];

  const rules = {
    mode: {
      required: true,
      validator: (value: string, cb: any) => {
        if (!/^0?[0-7]{3}$/.test(value)) {
          cb(t('app.file.modeDrawer.modeError'));
        }
        cb();
      },
    },
    user: { required: true, message: t('app.file.modeDrawer.userRequired') },
    group: { required: true, message: t('app.file.modeDrawer.groupRequired') },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

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
    () => [formState.ownerAccess, formState.groupAccess, formState.otherAccess],
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
      formState.ownerAccess = calculateAccess(owner);
      formState.groupAccess = calculateAccess(group);
      formState.otherAccess = calculateAccess(others);
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
    formState.ownerAccess = calculateAccess(mode.charAt(1));
    formState.groupAccess = calculateAccess(mode.charAt(2));
    formState.otherAccess = calculateAccess(mode.charAt(3));
  };

  const handleOk = () => {};
  const handleCancel = () => {
    visible.value = false;
  };

  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
