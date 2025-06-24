import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type {
  PermissionConfig,
  PermissionAccess,
  ParsedPermission,
} from '../types';

// 默认权限配置
const DEFAULT_PERMISSION: PermissionConfig = {
  mode: '0644',
  user: 'root',
  group: 'root',
  description: '',
};

export function usePermission(initialValue?: string) {
  const { t } = useI18n();

  // 当前权限配置
  const permission = ref<PermissionConfig>({ ...DEFAULT_PERMISSION });

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

  // 生成友好的权限描述
  const generateDescription = (
    mode: string,
    user: string,
    group: string
  ): string => {
    if (!/^0?[0-7]{3,4}$/.test(mode)) {
      return t('app.logrotate.permission.invalid_mode');
    }

    const paddedMode = mode.padStart(4, '0');
    const [, owner, groupDigit, other] = paddedMode.split('');

    const getPermissionText = (digit: string): string => {
      const n = parseInt(digit, 10);
      const perms = [];
      if (n & 4) perms.push(t('app.logrotate.permission.read'));
      if (n & 2) perms.push(t('app.logrotate.permission.write'));
      if (n & 1) perms.push(t('app.logrotate.permission.execute'));
      return perms.length > 0
        ? perms.join('、')
        : t('app.logrotate.permission.no_permission');
    };

    const ownerPerms = getPermissionText(owner);
    const groupPerms = getPermissionText(groupDigit);
    const otherPerms = getPermissionText(other);

    return t('app.logrotate.permission.friendly_format', {
      user,
      group,
      ownerPerms,
      groupPerms,
      otherPerms,
    });
  };

  // 解析create字符串
  const parseCreateString = (value: string): ParsedPermission => {
    if (!value) {
      return {
        isValid: false,
        mode: DEFAULT_PERMISSION.mode,
        user: DEFAULT_PERMISSION.user,
        group: DEFAULT_PERMISSION.group,
        access: { owner: [], group: [], other: [] },
      };
    }

    const match = value.match(/(?:create\s+)?(\d{3,4})\s+(\w+)\s+(\w+)/);
    if (!match) {
      return {
        isValid: false,
        mode: DEFAULT_PERMISSION.mode,
        user: DEFAULT_PERMISSION.user,
        group: DEFAULT_PERMISSION.group,
        access: { owner: [], group: [], other: [] },
      };
    }

    const [, mode, user, group] = match;
    const paddedMode = mode.padStart(4, '0');
    const [, owner, groupDigit, other] = paddedMode.split('');

    return {
      isValid: true,
      mode: paddedMode,
      user,
      group,
      access: {
        owner: calculateAccess(owner),
        group: calculateAccess(groupDigit),
        other: calculateAccess(other),
      },
    };
  };

  // 从字符串更新权限
  const updateFromString = (value: string): void => {
    const parsed = parseCreateString(value);
    if (parsed.isValid) {
      permission.value = {
        mode: parsed.mode,
        user: parsed.user,
        group: parsed.group,
        description: generateDescription(
          parsed.mode,
          parsed.user,
          parsed.group
        ),
      };
    } else {
      // 当输入无效时，设置默认权限并生成描述
      permission.value = {
        mode: DEFAULT_PERMISSION.mode,
        user: DEFAULT_PERMISSION.user,
        group: DEFAULT_PERMISSION.group,
        description: generateDescription(
          DEFAULT_PERMISSION.mode,
          DEFAULT_PERMISSION.user,
          DEFAULT_PERMISSION.group
        ),
      };
    }
  };

  // 从权限访问更新模式
  const updateFromAccess = (access: PermissionAccess): void => {
    const owner = calculateMode(access.owner);
    const group = calculateMode(access.group);
    const other = calculateMode(access.other);
    const mode = `0${owner}${group}${other}`;

    permission.value.mode = mode;
    permission.value.description = generateDescription(
      mode,
      permission.value.user,
      permission.value.group
    );
  };

  // 转换为字符串
  const toString = (config?: PermissionConfig): string => {
    const target = config || permission.value;
    return `create ${target.mode} ${target.user} ${target.group}`;
  };

  // 获取权限访问对象
  const getAccess = computed((): PermissionAccess => {
    const mode = permission.value.mode;
    if (!/^0?[0-7]{3,4}$/.test(mode)) {
      return { owner: [], group: [], other: [] };
    }

    const paddedMode = mode.padStart(4, '0');
    const [, owner, group, other] = paddedMode.split('');

    return {
      owner: calculateAccess(owner),
      group: calculateAccess(group),
      other: calculateAccess(other),
    };
  });

  // 初始化
  if (initialValue) {
    updateFromString(initialValue);
  } else {
    // 没有初始值时，生成默认权限描述
    permission.value.description = generateDescription(
      DEFAULT_PERMISSION.mode,
      DEFAULT_PERMISSION.user,
      DEFAULT_PERMISSION.group
    );
  }

  return {
    permission: computed(() => permission.value),
    access: getAccess,
    updateFromString,
    updateFromAccess,
    toString,
    parseCreateString,
  };
}
