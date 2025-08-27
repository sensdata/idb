<script lang="tsx">
  import { defineComponent, ref, h, compile, computed, inject } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter, RouteRecordRaw } from 'vue-router';
  import type { RouteMeta } from 'vue-router';
  import { useAppStore } from '@/store';
  import { listenerRouteChange } from '@/utils/route-listener';
  import { openWindow, regexUrl } from '@/utils';
  import usePermission from '@/composables/permission';
  import HostInfo from '@/components/host-info/index.vue';
  import useMenuTree from './use-menu-tree';

  export default defineComponent({
    emit: ['collapse'],
    setup() {
      const { t } = useI18n();
      const appStore = useAppStore();
      const router = useRouter();
      const route = useRoute();
      const permission = usePermission();
      const { manageMenuTree, appMenuTree } = useMenuTree();
      const isAppRoute = computed(() => route.fullPath.startsWith('/app'));
      const menuTree = computed(() => {
        if (isAppRoute.value) {
          return appMenuTree.value;
        }
        return manageMenuTree.value;
      });
      const collapsed = computed({
        get() {
          if (appStore.device === 'desktop') return appStore.menuCollapse;
          return false;
        },
        set(value: boolean) {
          appStore.updateSettings({ menuCollapse: value });
        },
      });

      const topMenu = computed(() => appStore.topMenu);
      const openKeys = ref<string[]>([]);
      const selectedKey = ref<string[]>([]);

      const openTerminal = inject<() => void>('openTerminal');

      const goto = (item: RouteRecordRaw) => {
        if (item.meta?.command === 'openTerminal') {
          openTerminal?.();
          return;
        }
        // Open external link
        if (regexUrl.test(item.path)) {
          openWindow(item.path);
          selectedKey.value = [item.name as string];
          return;
        }
        // Eliminate external link side effects
        const { hideInMenu, activeMenu } = item.meta as RouteMeta;
        if (route.name === item.name && !hideInMenu && !activeMenu) {
          selectedKey.value = [item.name as string];
          return;
        }
        // Trigger router change
        router.push({
          name: item.name,
        });
      };
      const findMenuOpenKeys = (
        target: string,
        targetMenuTree?: RouteRecordRaw[]
      ) => {
        const result: string[] = [];
        let isFind = false;
        const searchTree = targetMenuTree || menuTree.value;
        const backtrack = (item: RouteRecordRaw, keys: string[]) => {
          if (item.name === target) {
            isFind = true;
            result.push(...keys);
            return;
          }
          if (item.children?.length) {
            item.children.forEach((el) => {
              backtrack(el, [...keys, el.name as string]);
            });
          }
        };
        searchTree.forEach((el: RouteRecordRaw) => {
          if (isFind) return; // Performance optimization
          backtrack(el, [el.name as string]);
        });
        return result;
      };

      // 检查当前路由是否是菜单项的子路由
      const isChildOfMenu = (menuName: string) => {
        const menuOpenKeys = findMenuOpenKeys(route.name as string);
        return menuOpenKeys.includes(menuName);
      };

      // 查找菜单项的第一个子路由
      const findFirstChildRoute = (menuName: string): RouteRecordRaw | null => {
        let firstChild: RouteRecordRaw | null = null;

        const findChild = (items: RouteRecordRaw[]) => {
          for (const item of items) {
            if (item.name === menuName && item.children?.length) {
              // 查找第一个可以导航到的有效子路由
              for (const child of item.children) {
                if (!child.meta?.hideInMenu && permission.accessRouter(child)) {
                  firstChild = child;
                  return;
                }
              }
            } else if (item.children?.length) {
              findChild(item.children);
            }
          }
        };

        findChild(menuTree.value);
        return firstChild;
      };

      // 处理菜单项点击事件
      const handleMenuItemClick = (key: string) => {
        // 查找对应的菜单项
        const findMenuItem = (
          items: RouteRecordRaw[]
        ): RouteRecordRaw | null => {
          for (const item of items) {
            if (item.name === key) {
              return item;
            }
            if (item.children?.length) {
              const found = findMenuItem(item.children);
              if (found) return found;
            }
          }
          return null;
        };

        const menuItem = findMenuItem(menuTree.value);
        if (menuItem) {
          goto(menuItem);
        }
      };

      // 处理子菜单点击事件 - 当点击子菜单时
      const handleSubMenuClick = (key: string) => {
        // 如果菜单已展开，点击会自动收起（默认行为）
        if (openKeys.value.includes(key)) {
          return;
        }

        // 如果用户不在当前菜单的任意子页面，导航到第一个子页面
        if (!isChildOfMenu(key)) {
          const firstChild = findFirstChildRoute(key);
          if (firstChild) {
            goto(firstChild);
          }
        }
        // 如果用户已在子页面，只需展开菜单（默认行为）
      };

      // 处理子菜单展开/折叠状态变化
      const updateOpenKeys = (keys: string[]) => {
        // 找出新增的 key（即刚被打开的子菜单）
        const newKey = keys.find((key) => !openKeys.value.includes(key));
        if (newKey) {
          // 新展开的菜单
          handleSubMenuClick(newKey);
        }
        // 更新打开的菜单键值
        openKeys.value = keys;
      };

      listenerRouteChange((newRoute) => {
        const { requiresAuth, activeMenu, hideInMenu } = newRoute.meta;
        if (requiresAuth && (!hideInMenu || activeMenu)) {
          // 根据新路由确定应该使用哪个菜单树
          const isNewRouteApp = newRoute.fullPath.startsWith('/app');
          const targetMenuTree = isNewRouteApp
            ? appMenuTree.value
            : manageMenuTree.value;

          const menuOpenKeys = findMenuOpenKeys(
            (activeMenu || newRoute.name) as string,
            targetMenuTree
          );

          const keySet = new Set([...menuOpenKeys, ...openKeys.value]);
          openKeys.value = [...keySet];

          selectedKey.value = [
            activeMenu || menuOpenKeys[menuOpenKeys.length - 1],
          ];
        }
      }, true);
      const setCollapse = (val: boolean) => {
        if (appStore.device === 'desktop')
          appStore.updateSettings({ menuCollapse: val });
      };

      const renderSubMenu = () => {
        function travel(_route: RouteRecordRaw[], nodes = []) {
          if (_route) {
            _route.forEach((element) => {
              // This is demo, modify nodes as needed
              const icon = element?.meta?.icon
                ? () =>
                    typeof element?.meta?.icon === 'string' &&
                    element?.meta?.icon?.startsWith('icon')
                      ? h(compile(`<${element?.meta?.icon}/>`))
                      : h(compile(`${element?.meta?.icon}`))
                : null;
              const node =
                element?.children && element?.children.length !== 0 ? (
                  <a-sub-menu
                    key={element?.name}
                    v-slots={{
                      icon,
                      title: () =>
                        h(
                          'span',
                          { class: 'menu-title' },
                          t(element?.meta?.locale || '')
                        ),
                    }}
                  >
                    {travel(element?.children)}
                  </a-sub-menu>
                ) : (
                  <a-menu-item
                    key={element?.name}
                    v-slots={{ icon }}
                    onClick={() => goto(element)}
                  >
                    <span class="menu-title">
                      {t(element?.meta?.locale || '')}
                    </span>
                  </a-menu-item>
                );
              nodes.push(node as never);
            });
          }
          return nodes;
        }
        return travel(menuTree.value);
      };

      return () => (
        <div class="menu-container">
          {isAppRoute.value && (
            <div class="host-info-wrapper">
              <HostInfo collapsed={collapsed.value} />
            </div>
          )}
          <div class="menu-scroll">
            <a-menu
              mode={topMenu.value ? 'horizontal' : 'vertical'}
              v-model:collapsed={collapsed.value}
              v-model:open-keys={openKeys.value}
              show-collapse-button={appStore.device !== 'mobile'}
              auto-open
              selected-keys={selectedKey.value}
              auto-open-selected={true}
              level-indent={34}
              style={{ width: '100%' }}
              onCollapse={setCollapse}
              onClickSubMenu={handleSubMenuClick}
              onMenuItemClick={handleMenuItemClick}
              onUpdate:openKeys={updateOpenKeys}
            >
              {renderSubMenu()}
            </a-menu>
          </div>
        </div>
      );
    },
  });
</script>

<style scoped lang="less">
  @import '@/assets/style/mixin.less';

  .menu-container {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .host-info-wrapper {
    flex: 0 0 auto;
  }

  .menu-scroll {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overflow-x: visible;
    position: relative;

    :deep(.arco-menu) {
      .custom-scrollbar();

      // 确保子菜单能够正确显示
      .arco-menu-inline,
      .arco-menu-sub {
        position: relative;
        z-index: 1;
      }

      // 子菜单弹出层样式
      .arco-menu-inline-content {
        position: relative;
        z-index: 2;
      }
    }
  }

  :deep(.arco-menu) {
    height: 100%;
    overflow: hidden;

    // 折叠按钮样式
    .arco-menu-collapse-button {
      z-index: 2;
    }

    // 一级菜单项样式
    > .arco-menu-item,
    > .arco-menu-group-title,
    > .arco-menu-inline-header,
    > .arco-menu-pop-header {
      box-sizing: border-box;
      height: 40px;
      line-height: 40px;
      padding: 0 20px 0 48px !important; // 固定左侧内边距
      position: relative;

      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 3px;
        height: 100%;
        background-color: transparent;
      }
    }

    // 子菜单项样式 - 使用自然缩进
    .arco-menu-inline .arco-menu-item {
      box-sizing: border-box;
      height: 40px;
      line-height: 40px;
      position: relative;
      padding-right: 20px !important;
      display: flex;
      align-items: center;

      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 3px;
        height: 100%;
        background-color: transparent;
      }

      // 确保缩进列表和内容在同一行
      .arco-menu-indent-list {
        display: flex;
        align-items: center;
        flex-shrink: 0;
      }

      .arco-menu-item-inner {
        display: flex;
        align-items: center;
        flex: 1;
        min-width: 0;
      }
    }

    // 图标固定定位 - 只对一级菜单
    > .arco-menu-item > .arco-menu-icon,
    > .arco-menu-inline-header > .arco-menu-icon {
      position: absolute;
      left: 16px;
      top: 50%;
      transform: translateY(-50%);
      margin: 0;
    }

    // 菜单图标颜色 - 基础状态
    .arco-menu-icon svg {
      color: var(--color-text-2);
    }

    // 菜单项悬停状态图标颜色
    .arco-menu-item:hover .arco-menu-icon svg,
    .arco-menu-inline-header:hover .arco-menu-icon svg {
      color: var(--color-text-1);
    }

    // 菜单项选中状态图标颜色
    .arco-menu-item.arco-menu-selected .arco-menu-icon svg {
      color: rgb(var(--primary-6));
    }

    // 确保文本内容对齐
    .arco-menu-title,
    .arco-menu-item > span {
      display: block;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }

    // 统一菜单项文本样式
    .menu-title {
      display: inline-block;
      width: auto;
      text-align: left;
      padding-left: 0;
      position: relative;
      vertical-align: middle;
    }

    // 确保整个菜单项可点击，而不仅是标题区域
    .arco-menu-inline-header {
      cursor: pointer;
      width: 100%;

      &:hover {
        background-color: var(--color-fill-2);
      }
    }
  }
</style>
