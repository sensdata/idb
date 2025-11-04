import { ref } from 'vue';

/**
 * 判断 compose 名称是否为数据库类型
 */
export function useDatabaseManager() {
  const getDatabaseType = (
    composeName: string
  ): 'mysql' | 'postgresql' | 'redis' | null => {
    const name = composeName.toLowerCase();
    if (name.includes('mysql') || name.includes('mariadb')) return 'mysql';
    if (name.includes('postgresql') || name.includes('postgres'))
      return 'postgresql';
    if (name.includes('redis')) return 'redis';
    return null;
  };

  const isDatabaseCompose = (composeName: string): boolean => {
    return getDatabaseType(composeName) !== null;
  };

  return {
    getDatabaseType,
    isDatabaseCompose,
  };
}
