<template>
  <div class="port-rule-form">
    <a-form
      ref="formRef"
      :model="form"
      :rules="formRules"
      layout="vertical"
      @submit="handleSubmit"
    >
      <!-- 基本端口配置 -->
      <div class="form-section">
        <div class="section-header">
          <h3 class="section-title">{{
            $t('app.nftables.form.basicConfig')
          }}</h3>
        </div>

        <div class="form-content">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item
                field="port"
                :label="$t('app.nftables.config.rules.port')"
              >
                <a-input-number
                  v-model="form.port"
                  :min="1"
                  :max="65535"
                  :precision="0"
                  :placeholder="$t('app.nftables.config.rules.portPlaceholder')"
                  :readonly="props.portReadonly"
                  :disabled="props.portReadonly"
                  class="full-width"
                />
                <template #extra>
                  <div class="field-help">
                    {{ $t('app.nftables.form.portHelp') }}
                  </div>
                </template>
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item
                field="description"
                :label="$t('app.nftables.config.rules.description')"
              >
                <a-input
                  v-model="form.description"
                  :placeholder="$t('app.nftables.config.rules.descPlaceholder')"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
      </div>

      <!-- 规则配置模式选择 -->
      <div class="form-section">
        <div class="section-header">
          <h3 class="section-title">{{
            $t('app.nftables.form.configMode')
          }}</h3>
        </div>

        <div class="form-content">
          <div class="mode-selector">
            <div
              class="mode-card"
              :class="{ active: configMode === 'simple' }"
              @click="configMode = 'simple'"
            >
              <div class="mode-radio">
                <a-radio v-model="configMode" value="simple" />
              </div>
              <div class="mode-content">
                <div class="mode-title">
                  {{ $t('app.nftables.form.simpleMode') }}
                </div>
                <div class="mode-desc">
                  {{ $t('app.nftables.form.simpleModeDesc') }}
                </div>
              </div>
            </div>

            <div
              class="mode-card"
              :class="{ active: configMode === 'advanced' }"
              @click="configMode = 'advanced'"
            >
              <div class="mode-radio">
                <a-radio v-model="configMode" value="advanced" />
              </div>
              <div class="mode-content">
                <div class="mode-title">
                  {{ $t('app.nftables.form.advancedMode') }}
                </div>
                <div class="mode-desc">
                  {{ $t('app.nftables.form.advancedModeDesc') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 简单模式配置 -->
      <div v-if="configMode === 'simple'" class="form-section">
        <div class="section-header">
          <h3 class="section-title">
            {{ $t('app.nftables.form.accessControl') }}
          </h3>
        </div>

        <div class="form-content">
          <a-form-item
            field="simpleAction"
            :label="$t('app.nftables.config.rules.action')"
          >
            <div class="action-selector">
              <div
                class="action-card accept"
                :class="{ active: form.simpleAction === 'accept' }"
                @click="form.simpleAction = 'accept'"
              >
                <div class="action-radio">
                  <a-radio v-model="form.simpleAction" value="accept" />
                </div>
                <div class="action-content">
                  <icon-check-circle class="action-icon" />
                  <div class="action-info">
                    <div class="action-title">
                      {{ $t('app.nftables.config.rules.allow') }}
                    </div>
                    <div class="action-desc">
                      {{ $t('app.nftables.form.allowDesc') }}
                    </div>
                  </div>
                </div>
              </div>

              <div
                class="action-card drop"
                :class="{ active: form.simpleAction === 'drop' }"
                @click="form.simpleAction = 'drop'"
              >
                <div class="action-radio">
                  <a-radio v-model="form.simpleAction" value="drop" />
                </div>
                <div class="action-content">
                  <icon-close-circle class="action-icon" />
                  <div class="action-info">
                    <div class="action-title">
                      {{ $t('app.nftables.config.rules.deny') }}
                    </div>
                    <div class="action-desc">
                      {{ $t('app.nftables.form.denyDesc') }}
                    </div>
                  </div>
                </div>
              </div>

              <div
                class="action-card reject"
                :class="{ active: form.simpleAction === 'reject' }"
                @click="form.simpleAction = 'reject'"
              >
                <div class="action-radio">
                  <a-radio v-model="form.simpleAction" value="reject" />
                </div>
                <div class="action-content">
                  <icon-minus-circle class="action-icon" />
                  <div class="action-info">
                    <div class="action-title">REJECT</div>
                    <div class="action-desc">
                      {{ $t('app.nftables.form.rejectDesc') }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </a-form-item>
        </div>
      </div>

      <!-- 高级模式配置 -->
      <div v-else-if="configMode === 'advanced'" class="form-section">
        <div class="section-header">
          <h3 class="section-title">{{ $t('app.nftables.form.rulesList') }}</h3>
          <a-button type="primary" size="small" @click="addRule">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.nftables.form.addRule') }}
          </a-button>
        </div>

        <div class="form-content">
          <div v-if="form.rules.length === 0" class="empty-state">
            <a-empty :description="$t('app.nftables.form.noRulesHint')">
              <template #image>
                <icon-file class="empty-icon" />
              </template>
              <a-button type="primary" @click="addRule">
                {{ $t('app.nftables.form.addFirstRule') }}
              </a-button>
            </a-empty>
          </div>

          <div v-else class="rules-list">
            <div
              v-for="(rule, index) in form.rules"
              :key="index"
              class="rule-card"
            >
              <div class="rule-header">
                <div class="rule-title">
                  <span class="rule-number"
                    >{{ $t('app.nftables.form.rule') }} {{ index + 1 }}</span
                  >
                  <a-tag :color="getRuleTypeColor(rule.type)" size="small">
                    {{ getRuleTypeLabel(rule.type) }}
                  </a-tag>
                </div>
                <a-button
                  type="text"
                  size="small"
                  status="danger"
                  @click="removeRule(index)"
                >
                  <template #icon>
                    <icon-delete />
                  </template>
                </a-button>
              </div>

              <div class="rule-content">
                <a-row :gutter="16">
                  <a-col :span="8">
                    <a-form-item
                      :field="`rules.${index}.type`"
                      :label="$t('app.nftables.form.ruleType')"
                    >
                      <a-select
                        v-model="rule.type"
                        @change="handleRuleTypeChange(rule)"
                      >
                        <a-option value="default">
                          <div class="rule-type-option">
                            <icon-check class="option-icon" />
                            {{ $t('app.nftables.form.basicRule') }}
                          </div>
                        </a-option>
                        <a-option value="rate_limit">
                          <div class="rule-type-option">
                            <icon-clock-circle class="option-icon" />
                            {{ $t('app.nftables.form.rateLimit') }}
                          </div>
                        </a-option>
                        <a-option value="concurrent_limit">
                          <div class="rule-type-option">
                            <icon-user-group class="option-icon" />
                            {{ $t('app.nftables.form.concurrentLimit') }}
                          </div>
                        </a-option>
                      </a-select>
                    </a-form-item>
                  </a-col>

                  <a-col :span="8">
                    <a-form-item
                      :field="`rules.${index}.action`"
                      :label="$t('app.nftables.config.rules.action')"
                    >
                      <a-select v-model="rule.action">
                        <a-option value="accept">
                          <a-tag color="green" size="small">
                            {{ $t('app.nftables.config.rules.allow') }}
                          </a-tag>
                        </a-option>
                        <a-option value="drop">
                          <a-tag color="red" size="small">
                            {{ $t('app.nftables.config.rules.deny') }}
                          </a-tag>
                        </a-option>
                        <a-option value="reject">
                          <a-tag color="orange" size="small"> REJECT </a-tag>
                        </a-option>
                      </a-select>
                    </a-form-item>
                  </a-col>

                  <a-col :span="8">
                    <a-form-item
                      v-if="rule.type === 'rate_limit'"
                      :field="`rules.${index}.rate`"
                      :label="$t('app.nftables.form.rateValue')"
                    >
                      <a-select v-model="rule.rate" allow-create>
                        <a-option value="10/second">10/second</a-option>
                        <a-option value="100/second">100/second</a-option>
                        <a-option value="1000/second">1000/second</a-option>
                        <a-option value="10/minute">10/minute</a-option>
                        <a-option value="100/minute">100/minute</a-option>
                      </a-select>
                      <template #extra>
                        <div class="field-help">
                          {{ $t('app.nftables.form.rateHelpText') }}
                        </div>
                      </template>
                    </a-form-item>
                    <a-form-item
                      v-else-if="rule.type === 'concurrent_limit'"
                      :field="`rules.${index}.count`"
                      :label="$t('app.nftables.form.concurrentCount')"
                    >
                      <a-input-number
                        v-model="rule.count"
                        :min="1"
                        :max="10000"
                        :precision="0"
                        :placeholder="
                          $t('app.nftables.form.concurrentPlaceholder')
                        "
                        class="full-width"
                      />
                      <template #extra>
                        <div class="field-help">
                          {{ $t('app.nftables.form.concurrentHelpText') }}
                        </div>
                      </template>
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 预览配置 -->
      <div class="form-section">
        <div class="section-header">
          <h3 class="section-title">
            {{ $t('app.nftables.form.configPreview') }}
          </h3>
        </div>
        <div class="form-content">
          <div class="config-preview">
            <pre>{{ generateConfigPreview() }}</pre>
          </div>
        </div>
      </div>

      <!-- 表单操作按钮 -->
      <div class="form-actions">
        <a-space size="large">
          <a-button size="large" @click="handleCancel">
            {{ $t('app.nftables.form.cancel') }}
          </a-button>

          <a-button
            type="primary"
            html-type="submit"
            :loading="loading"
            size="large"
          >
            {{
              editingRule
                ? $t('app.nftables.form.updateRule')
                : $t('app.nftables.button.addRule')
            }}
          </a-button>
        </a-space>
      </div>
    </a-form>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, nextTick, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    IconPlus,
    IconDelete,
    IconSave,
    IconClose,
    IconCheckCircle,
    IconCloseCircle,
    IconMinusCircle,
    IconCheck,
    IconClockCircle,
    IconUserGroup,
    IconFile,
  } from '@arco-design/web-vue/es/icon';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import type { PortRule, RuleItem } from '@/api/nftables';

  // 表单专用类型
  interface PortRuleForm {
    port: number;
    description: string;
    simpleAction: 'accept' | 'drop' | 'reject'; // 简单模式的动作
    rules: RuleItem[]; // 高级规则列表
  }

  interface Props {
    loading?: boolean;
    editingRule?: PortRule | null;
    portReadonly?: boolean;
  }

  interface Emits {
    (e: 'submit', rule: PortRule): void;
    (e: 'cancel'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    editingRule: null,
    portReadonly: false,
  });

  const emit = defineEmits<Emits>();
  const { t } = useI18n();
  const formRef = ref<FormInstance>();

  // 配置模式：simple（简单模式）或 advanced（高级模式）
  const configMode = ref<'simple' | 'advanced'>('simple');

  // 默认表单数据工厂函数
  const createDefaultForm = (): PortRuleForm => ({
    port: 80,
    description: '',
    simpleAction: 'accept',
    rules: [],
  });

  // 表单数据
  const form = reactive<PortRuleForm>(createDefaultForm());

  // 表单验证规则
  const formRules = computed(
    (): Record<string, FieldRule[]> => ({
      port: [
        {
          required: true,
          message: t('app.nftables.validation.portRequired'),
        },
        {
          type: 'number' as const,
          min: 1,
          max: 65535,
          message: t('app.nftables.validation.portRange'),
        },
      ],
      simpleAction: [
        {
          required: true,
          message: t('app.nftables.validation.actionRequired'),
        },
      ],
    })
  );

  // 重置表单
  const resetForm = () => {
    Object.assign(form, createDefaultForm());
    configMode.value = 'simple';

    nextTick(() => {
      formRef.value?.clearValidate();
    });
  };

  // 监听编辑规则变化
  watch(
    () => props.editingRule,
    (newRule) => {
      if (newRule) {
        // 确定配置模式
        const hasAdvancedRules = newRule.rules && newRule.rules.length > 0;
        configMode.value = hasAdvancedRules ? 'advanced' : 'simple';

        // 提取端口号（只取第一个端口）
        const portValue = Array.isArray(newRule.port)
          ? newRule.port[0]
          : newRule.port;

        // 设置表单数据
        Object.assign(form, {
          port: portValue,
          description: newRule.description || '',
          simpleAction: newRule.action || 'accept',
          rules: newRule.rules ? [...newRule.rules] : [],
        });
      } else {
        resetForm();
      }
    },
    { immediate: true }
  );

  // 获取规则类型标签
  const getRuleTypeLabel = (type: string): string => {
    switch (type) {
      case 'default':
        return t('app.nftables.form.basicRule');
      case 'rate_limit':
        return t('app.nftables.form.rateLimit');
      case 'concurrent_limit':
        return t('app.nftables.form.concurrentLimit');
      default:
        return type;
    }
  };

  // 获取规则类型颜色
  const getRuleTypeColor = (type: string): string => {
    switch (type) {
      case 'default':
        return 'blue';
      case 'rate_limit':
        return 'orange';
      case 'concurrent_limit':
        return 'purple';
      default:
        return 'gray';
    }
  };

  // 规则类型变化处理
  const handleRuleTypeChange = (rule: RuleItem) => {
    // 重置规则特定字段
    if (rule.type !== 'rate_limit') {
      delete rule.rate;
    }
    if (rule.type !== 'concurrent_limit') {
      delete rule.count;
    }

    // 为新规则类型设置默认值
    if (rule.type === 'rate_limit' && !rule.rate) {
      rule.rate = '100/second';
    }
    if (rule.type === 'concurrent_limit' && !rule.count) {
      rule.count = 10;
    }
  };

  // 生成配置预览
  const generateConfigPreview = (): string => {
    if (configMode.value === 'simple') {
      return `tcp dport ${form.port} ${form.simpleAction}`;
    }

    if (form.rules.length === 0) {
      return t('app.nftables.form.noRulesPreview');
    }

    return form.rules
      .map((rule) => {
        switch (rule.type) {
          case 'rate_limit':
            return `tcp dport ${form.port} ip saddr limit rate ${
              rule.rate || '100/second'
            } ${rule.action}`;
          case 'concurrent_limit':
            return `tcp dport ${form.port} ct count ip saddr over ${
              rule.count || 10
            } ${rule.action}`;
          case 'default':
          default:
            return `tcp dport ${form.port} ${rule.action}`;
        }
      })
      .join('\n');
  };

  // 提交表单
  const handleSubmit = async ({ errors }: { errors?: any }) => {
    if (errors) return;

    let rules: RuleItem[] = [];

    if (configMode.value === 'simple') {
      // 简单模式：创建一个默认规则
      rules = [
        {
          type: 'default',
          action: form.simpleAction,
        },
      ];
    } else {
      // 高级模式：使用用户配置的规则
      rules =
        form.rules.length > 0
          ? [...form.rules]
          : [
              {
                type: 'default',
                action: 'accept',
              },
            ];
    }

    const ruleData: PortRule = {
      port: form.port,
      description: form.description || `TCP ${form.port}`,
      rules,
    };

    emit('submit', ruleData);
  };

  // 取消操作
  const handleCancel = () => {
    emit('cancel');
    resetForm();
  };

  // 添加高级规则
  const addRule = () => {
    const newRule: RuleItem = {
      type: 'default',
      action: 'accept',
    };
    form.rules.push(newRule);
  };

  // 删除高级规则
  const removeRule = (index: number) => {
    form.rules.splice(index, 1);
  };

  // 暴露方法供父组件调用
  defineExpose({
    resetForm,
    validateForm: () => formRef.value?.validate(),
  });
</script>

<style scoped lang="less">
  .port-rule-form {
    padding: 0;

    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: var(--color-text-1);
      margin: 0 0 16px 0;
      border-bottom: 2px solid var(--color-primary);
      padding-bottom: 8px;
    }

    .form-section {
      margin-bottom: 24px;

      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
      }

      .form-content {
        .mode-selector {
          display: flex;
          flex-direction: column;
          gap: 12px;

          .mode-card {
            display: flex;
            align-items: center;
            padding: 16px;
            border: 2px solid var(--color-border-2);
            border-radius: 8px;
            transition: all 0.2s;

            &:hover {
              border-color: var(--color-primary-light-3);
              background: var(--color-primary-light-5);
            }

            &.active {
              border-color: var(--color-primary);
              background: var(--color-primary-light-5);
            }

            .mode-radio {
              margin-right: 12px;
            }

            .mode-content {
              .mode-title {
                font-weight: 500;
                color: var(--color-text-1);
                margin-bottom: 4px;
              }

              .mode-desc {
                font-size: 12px;
                color: var(--color-text-3);
              }
            }
          }
        }
      }
    }

    .form-section {
      margin-bottom: 24px;
      padding: 20px;
      background: #fff;
      border: 1px solid var(--color-border-2);
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;

        .section-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--color-text-1);
          margin: 0;
        }
      }

      .form-content {
        .mode-selector {
          display: flex;
          flex-direction: column;
          gap: 12px;

          .mode-card {
            display: flex;
            align-items: center;
            padding: 16px;
            border: 1px solid var(--color-border-2);
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s ease;
            position: relative;

            &:hover {
              border-color: var(--color-primary-light-4);
              background: var(--color-fill-1);
              box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
            }

            &.active {
              border-color: var(--color-primary-light-4);
              background: var(--color-primary-light-5);
              box-shadow: 0 0 0 2px var(--color-primary-light-4);
            }

            .mode-radio {
              margin-right: 12px;
              pointer-events: none;

              :deep(.arco-radio) {
                margin: 0;
              }
            }

            .mode-content {
              flex: 1;

              .mode-title {
                font-weight: 500;
                color: var(--color-text-1);
                margin-bottom: 4px;
                font-size: 14px;
              }

              .mode-desc {
                font-size: 12px;
                color: var(--color-text-3);
                line-height: 1.4;
              }
            }
          }
        }

        .action-selector {
          display: flex;
          flex-direction: column;
          gap: 12px;

          .action-card {
            display: flex;
            align-items: center;
            padding: 16px;
            border: 1px solid var(--color-border-2);
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s ease;
            position: relative;

            &:hover {
              border-color: var(--color-primary-light-4);
              background: var(--color-fill-1);
              box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
            }

            &.active {
              border-color: var(--color-primary-light-4);
              background: var(--color-primary-light-5);
              box-shadow: 0 0 0 2px var(--color-primary-light-4);
            }

            .action-radio {
              margin-right: 12px;
              pointer-events: none;

              :deep(.arco-radio) {
                margin: 0;
              }
            }

            .action-content {
              display: flex;
              align-items: center;
              flex: 1;

              .action-icon {
                font-size: 20px;
                margin-right: 12px;
              }

              .action-info {
                .action-title {
                  font-weight: 500;
                  color: var(--color-text-1);
                  margin-bottom: 4px;
                  font-size: 14px;
                }

                .action-desc {
                  font-size: 12px;
                  color: var(--color-text-3);
                  line-height: 1.4;
                }
              }
            }

            &.accept .action-icon {
              color: var(--color-success);
            }

            &.drop .action-icon {
              color: var(--color-danger);
            }

            &.reject .action-icon {
              color: var(--color-warning);
            }
          }
        }

        .empty-state {
          padding: 40px 20px;
          text-align: center;
          background: var(--color-fill-1);
          border-radius: 8px;
          border: 2px dashed var(--color-border-2);

          .empty-icon {
            font-size: 48px;
            color: var(--color-text-3);
          }
        }

        .rules-list {
          .rule-card {
            margin-bottom: 16px;
            padding: 16px;
            border: 1px solid var(--color-border-2);
            border-radius: 8px;
            background: var(--color-fill-1);

            &:last-child {
              margin-bottom: 0;
            }

            .rule-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 16px;

              .rule-title {
                display: flex;
                align-items: center;
                gap: 8px;

                .rule-number {
                  font-weight: 500;
                  color: var(--color-text-1);
                }
              }
            }

            .rule-content {
              .rule-type-option {
                display: flex;
                align-items: center;
                gap: 8px;

                .option-icon {
                  font-size: 14px;
                }
              }
            }
          }
        }

        .config-preview {
          background: var(--color-fill-1);
          border: 1px solid var(--color-border-2);
          border-radius: 6px;
          padding: 16px;

          pre {
            margin: 0;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 12px;
            color: var(--color-text-2);
            white-space: pre-wrap;
            word-break: break-all;
          }
        }
      }
    }

    .field-help {
      font-size: 12px;
      color: var(--color-text-3);
      margin-top: 4px;
      line-height: 1.4;
    }

    .full-width {
      width: 100%;
    }

    .form-actions {
      display: flex;
      justify-content: flex-end;
      padding: 20px;
      margin-top: 8px;
      border-top: 1px solid var(--color-border-2);
      background: var(--color-fill-1);
    }

    :deep(.arco-form-item-label) {
      font-weight: 500;
      color: var(--color-text-1);
    }

    :deep(.arco-input-number) {
      width: 100%;
    }

    :deep(.arco-select) {
      width: 100%;
    }

    // 响应式设计
    @media (max-width: 768px) {
      .form-section {
        padding: 16px;
        margin-bottom: 16px;
      }

      .form-content {
        .mode-selector .mode-card,
        .action-selector .action-card {
          padding: 12px;
        }
      }

      .form-actions {
        padding: 16px;

        :deep(.arco-space) {
          width: 100%;
          justify-content: flex-end;
        }
      }
    }
  }
</style>
