<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.createFolderDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :rules="rules"
      :labelAlign="'left'"
    >
      <a-form-item
        field="name"
        :label="$t('components.file.createFolderDrawer.name')"
        label-col-flex="90px"
      >
        <a-input v-model="formState.name" />
      </a-form-item>
      <a-form-item field="set_mode" label=" " label-col-flex="90px">
        <a-checkbox v-model="formState.set_mode">
          {{ $t('components.file.createFolderDrawer.set_mode') }}
        </a-checkbox>
      </a-form-item>
      <template v-if="formState.set_mode">
        <a-form-item
          field="owner_access"
          :label="$t('components.file.createFolderDrawer.owner_access')"
          label-col-flex="90px"
        >
          <a-checkbox-group
            v-model="formState.owner_access"
            :options="accessOptions"
          />
        </a-form-item>
        <a-form-item
          field="group_access"
          :label="$t('components.file.createFolderDrawer.group_access')"
          label-col-flex="90px"
        >
          <a-checkbox-group
            v-model="formState.group_access"
            :options="accessOptions"
          />
        </a-form-item>
        <a-form-item
          field="other_access"
          :label="$t('components.file.createFolderDrawer.other_access')"
          label-col-flex="90px"
        >
          <a-checkbox-group
            v-model="formState.other_access"
            :options="accessOptions"
          />
        </a-form-item>
        <a-form-item
          field="mode"
          :label="$t('components.file.createFolderDrawer.mode')"
          label-col-flex="90px"
        >
          <a-input v-model="formState.mode" class="w-60" :max-length="4" />
        </a-form-item>
      </template>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { createFileApi } from '@/api/file';
  import { useHostStore } from '@/store';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const hostStore = useHostStore();
  const formRef = ref();
  const formState = reactive({
    name: '',
    pwd: '',
    set_mode: false,
    mode: '0755',
    owner_access: ['4', '2', '1'],
    group_access: ['4', '1'],
    other_access: ['4', '1'],
  });

  const rules = {
    name: {
      required: true,
      message: t('components.file.createFolderDrawer.name_required'),
    },
    mode: {
      required: true,
      validator: (value: string, cb: any) => {
        if (!/^0?[0-7]{3}$/.test(value)) {
          cb(t('components.file.createFolderDrawer.modeError'));
        }
        cb();
      },
    },
  };

  const accessOptions = [
    { label: t('components.file.modeDrawer.read'), value: '4' },
    { label: t('components.file.modeDrawer.write'), value: '2' },
    { label: t('components.file.modeDrawer.execute'), value: '1' },
  ];
  const calculateMode = (access: string[]) => {
    return access.reduce((sum, per) => sum + Number(per), 0);
  };

  const calculateAccess = (digit: string) => {
    const arr: string[] = [];
    const n = parseInt(digit, 10);
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

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const setData = (data: { pwd: string; mode?: string }) => {
    formState.pwd = data.pwd;
    const mode = data.mode || '0755';
    formState.mode = mode.padStart(4, '0');
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    if (!(await validate())) {
      return;
    }

    if (loading.value) {
      return;
    }
    setLoading(true);
    try {
      await createFileApi({
        host: hostStore.currentId ?? hostStore.defaultId,
        source: formState.pwd + '/' + formState.name,
        is_dir: true,
        ...(formState.set_mode
          ? {
              mode: +formState.mode,
            }
          : {}),
      });
      visible.value = false;
      Message.success(t('components.file.createFolderDrawer.success'));
      emit('ok');
    } catch (err: any) {
      Message.error(err.message);
    } finally {
      setLoading(false);
    }
  };
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
