/**
 * 地址栏组件相关类型定义
 */

/**
 * 事件发射函数类型
 */
export type EmitFn = {
  (event: 'goto', path: string): void;
  (event: 'clear'): void;
  (event: 'search', payload: { path: string; word?: string }): void;
};
