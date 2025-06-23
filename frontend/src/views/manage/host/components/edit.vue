<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.form.title.edit')"
    width="540px"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="model" :rules="rules">
      <a-form-item field="name" :label="$t('manage.host.form.name.label')">
        <a-input
          v-model="model.name"
          :placeholder="$t('manage.host.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item field="group" :label="$t('manage.host.form.group.label')">
        <a-select
          v-model="model.group_id"
          :placeholder="$t('manage.host.form.group.placeholder')"
          :loading="groupLoading"
          :options="groupOptions"
          allow-clear
        >
          <template #footer>
            <a-link class="group-add" @click="handleAddGroup">
              <icon-plus class="icon-plus" />
              {{ $t('manage.host.form.group.add') }}
            </a-link>
          </template>
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>
  <group-form ref="groupFormRef" @ok="handleGroupFormOk" />
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import { updateHostApi, getHostGroupListApi } from '@/api/host';
  import useVisible from '@/composables/visible';
  import useLoading from '@/composables/loading';
  import type { HostEntity } from '@/entity/Host';
  import GroupForm from './group-form.vue';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const model = reactive({
    id: 0,
    name: '',
    group_id: 0,
  });

  const rules = computed(() => ({
    name: [{ required: true, message: t('manage.host.form.name.required') }],
  }));

  const groupOptions = ref<SelectOption[]>([]);

  const getData = () => {
    const data = toRaw(model);
    return data;
  };

  const reset = () => {
    formRef.value.resetFields();
    formRef.value.clearValidate();
  };

  const { loading: groupLoading, setLoading: setGroupLoading } = useLoading();
  const loadGroupOptions = async () => {
    setGroupLoading(true);
    try {
      const ret = await getHostGroupListApi({
        page: 1,
        page_size: 1000,
      });
      groupOptions.value = [
        {
          label: t('manage.host.form.group.default'),
          value: 0,
        },
        ...ret.items.map((item: any) => ({
          label: item.group_name,
          value: item.id,
        })),
      ];
    } catch (err: any) {
      // Message.error(err?.message);
    } finally {
      setGroupLoading(false);
    }
  };

  const loadOptions = async () => {
    await loadGroupOptions();
  };

  const groupFormRef = ref<InstanceType<typeof GroupForm>>();

  const handleAddGroup = () => {
    const form = groupFormRef.value;
    form?.reset();
    form?.show();
  };

  const handleGroupFormOk = () => {
    loadGroupOptions();
  };

  const setData = (data: HostEntity) => {
    model.id = data.id;
    model.name = data.name;
    model.group_id = data.group?.id || 0;
  };

  const validate = () => {
    return formRef.value.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      const data = getData();
      await updateHostApi(data);
      Message.success(t('manage.host.form.save.success'));
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
    reset,
    loadOptions,
    setData,
  });
</script>

<style scoped lang="less">
  .group-add {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 8px 12px;
    .icon-plus {
      margin-right: 4px;
    }
  }
</style>
