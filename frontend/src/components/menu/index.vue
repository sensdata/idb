<script lang="tsx">
  import { defineComponent, ref, h, compile, computed, inject } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter, RouteRecordRaw } from 'vue-router';
  import type { RouteMeta } from 'vue-router';
  import { useAppStore } from '@/store';
  import { listenerRouteChange } from '@/utils/route-listener';
  import { openWindow, regexUrl } from '@/utils';
  import usePermission from '@/hooks/permission';
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
      const findMenuOpenKeys = (target: string) => {
        const result: string[] = [];
        let isFind = false;
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
        menuTree.value.forEach((el: RouteRecordRaw) => {
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
          const menuOpenKeys = findMenuOpenKeys(
            (activeMenu || newRoute.name) as string
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
        <>
          {isAppRoute.value && (
            <HostInfo collapsed={collapsed.value} class="mb-2" />
          )}
          <a-menu
            mode={topMenu.value ? 'horizontal' : 'vertical'}
            v-model:collapsed={collapsed.value}
            v-model:open-keys={openKeys.value}
            show-collapse-button={appStore.device !== 'mobile'}
            auto-open
            selected-keys={selectedKey.value}
            auto-open-selected={true}
            level-indent={34}
            style={{ width: '100%', flex: 1 }}
            onCollapse={setCollapse}
            onClickSubMenu={handleSubMenuClick}
            onMenuItemClick={handleMenuItemClick}
            onUpdate:openKeys={updateOpenKeys}
          >
            {renderSubMenu()}
          </a-menu>
        </>
      );
    },
  });
</script>

<style scoped lang="less">
  :deep(.arco-menu) {
    // 通用样式重置
    .arco-menu-item,
    .arco-menu-group-title,
    .arco-menu-inline-header,
    .arco-menu-pop-header {
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

    // 图标固定定位
    .arco-menu-icon,
    .arco-menu-item > .arco-icon,
    .arco-menu-inline-header > .arco-icon {
      position: absolute;
      left: 16px;
      top: 50%;
      transform: translateY(-50%);
      margin: 0;
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

    // 降低箭头图标的优先级，避免干扰对齐
    .arco-icon-down,
    .arco-icon-right {
      position: absolute;
      right: 16px;
      left: auto;
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
