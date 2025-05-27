<template>
  <a-modal
    :visible="props.visible"
    :title="$t('app.logrotate.permission.modal_title')"
    :width="560"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <div class="permission-modal-content">
      <!-- 权限设置区域 -->
      <div class="permission-section">
        <div class="section-title">
          {{ $t('app.logrotate.permission.title') }}
        </div>

        <!-- 权限复选框 -->
        <div class="permission-checkboxes">
          <a-row :gutter="16">
            <a-col :span="8">
              <div class="permission-group">
                <div class="permission-label">
                  {{ $t('app.logrotate.permission.owner') }}
                </div>
                <a-checkbox-group
                  v-model="tempAccess.owner"
                  :options="accessOptions"
                  @change="updateModeFromCheckboxes"
                />
              </div>
            </a-col>
            <a-col :span="8">
              <div class="permission-group">
                <div class="permission-label">
                  {{ $t('app.logrotate.permission.group') }}
                </div>
                <a-checkbox-group
                  v-model="tempAccess.group"
                  :options="accessOptions"
                  @change="updateModeFromCheckboxes"
                />
              </div>
            </a-col>
            <a-col :span="8">
              <div class="permission-group">
                <div class="permission-label">
                  {{ $t('app.logrotate.permission.other') }}
                </div>
                <a-checkbox-group
                  v-model="tempAccess.other"
                  :options="accessOptions"
                  @change="updateModeFromCheckboxes"
                />
              </div>
            </a-col>
          </a-row>
        </div>

        <!-- 权限码输入 -->
        <div class="permission-mode">
          <a-form-item :label="$t('app.logrotate.permission.mode')">
            <a-input
              v-model="tempPermission.mode"
              :placeholder="$t('app.logrotate.permission.mode_placeholder')"
              style="width: 120px"
              @change="updateCheckboxesFromMode"
            />
          </a-form-item>
        </div>
      </div>

      <!-- 所有者设置区域 -->
      <div class="owner-section">
        <div class="section-title">
          {{ $t('app.logrotate.permission.ownership') }}
        </div>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item :label="$t('app.logrotate.permission.user')">
              <a-input
                v-model="tempPermission.user"
                :placeholder="$t('app.logrotate.permission.user_placeholder')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="$t('app.logrotate.permission.group_name')">
              <a-input
                v-model="tempPermission.group"
                :placeholder="$t('app.logrotate.permission.group_placeholder')"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { usePermission } from './hooks/use-permission';
  import type { PermissionConfig, PermissionAccess } from './types';

  interface Props {
    visible: boolean;
    permission: PermissionConfig;
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void;
    (e: 'confirm', permission: PermissionConfig): void;
  }

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  const { t } = useI18n();

  // 权限选项
  const accessOptions = computed(() => [
    { label: t('app.logrotate.permission.read'), value: '4' },
    { label: t('app.logrotate.permission.write'), value: '2' },
    { label: t('app.logrotate.permission.execute'), value: '1' },
  ]);

  // 临时权限状态
  const tempPermission = ref<PermissionConfig>({ ...props.permission });
  const tempAccess = ref<PermissionAccess>({
    owner: [],
    group: [],
    other: [],
  });

  // 使用权限hook进行转换
  const { parseCreateString } = usePermission();

  // 更新标志，防止循环更新
  let isUpdatingFromMode = false;
  let isUpdatingFromCheckboxes = false;

  // 计算权限值
  const calculateMode = (access: string[]): number => {
    return access.reduce((sum, per) => sum + Number(per), 0);
  };

  // 从数字计算权限数组
  const calculateAccess = (digit: string): string[] => {
    const arr: string[] = [];
    const n = parseInt(digit, 10);
    if (n & 4) arr.push('4');
    if (n & 2) arr.push('2');
    if (n & 1) arr.push('1');
    return arr;
  };

  // 从复选框更新权限码
  const updateModeFromCheckboxes = (): void => {
    if (isUpdatingFromMode) return;

    isUpdatingFromCheckboxes = true;
    const owner = calculateMode(tempAccess.value.owner);
    const group = calculateMode(tempAccess.value.group);
    const other = calculateMode(tempAccess.value.other);
    tempPermission.value.mode = `0${owner}${group}${other}`;
    isUpdatingFromCheckboxes = false;
  };

  // 从权限码更新复选框
  const updateCheckboxesFromMode = (): void => {
    if (isUpdatingFromCheckboxes) return;

    const mode = tempPermission.value.mode;
    if (!/^0?[0-7]{3,4}$/.test(mode)) return;

    isUpdatingFromMode = true;
    const paddedMode = mode.padStart(4, '0');
    const [, owner, group, other] = paddedMode.split('');

    tempAccess.value = {
      owner: calculateAccess(owner),
      group: calculateAccess(group),
      other: calculateAccess(other),
    };
    isUpdatingFromMode = false;
  };

  // 初始化临时状态
  const initTempState = (): void => {
    tempPermission.value = { ...props.permission };
    const createString = `create ${props.permission.mode} ${props.permission.user} ${props.permission.group}`;
    const parsed = parseCreateString(createString);
    if (parsed.isValid) {
      tempAccess.value = parsed.access;
    }
  };

  // 弹窗确认
  const handleOk = (): void => {
    emit('confirm', { ...tempPermission.value });
    emit('update:visible', false);
  };

  // 弹窗取消
  const handleCancel = (): void => {
    initTempState();
    emit('update:visible', false);
  };

  // 监听visible变化，重新初始化状态
  watch(
    () => props.visible,
    (newVisible) => {
      if (newVisible) {
        initTempState();
      }
    }
  );

  // 监听permission变化
  watch(
    () => props.permission,
    () => {
      if (props.visible) {
        initTempState();
      }
    },
    { deep: true }
  );
</script>

<style scoped>
  .permission-modal-content {
    max-height: 60vh;
    overflow: hidden auto;
  }

  .permission-section {
    margin-bottom: 24px;
  }

  .owner-section {
    margin-bottom: 16px;
  }

  .section-title {
    padding-bottom: 6px;
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
    border-bottom: 1px solid var(--color-border-3);
  }

  .permission-checkboxes {
    margin-bottom: 16px;
    overflow: hidden;
  }

  .permission-group {
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .permission-label {
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-2);
    text-align: left;
    word-wrap: break-word;
  }

  .permission-mode {
    display: flex;
    justify-content: center;
    overflow: hidden;
  }

  :deep(.arco-checkbox-group) {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  :deep(.arco-checkbox) {
    margin-right: 0;
  }

  :deep(.arco-modal-body) {
    overflow-x: hidden;
  }

  :deep(.arco-row) {
    margin-right: 0 !important;
    margin-left: 0 !important;
  }

  :deep(.arco-col) {
    padding-right: 8px;
    padding-left: 8px;
  }
</style>
