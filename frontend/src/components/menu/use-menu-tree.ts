import { computed } from 'vue';
import { RouteRecordRaw, RouteRecordNormalized } from 'vue-router';
import usePermission from '@/hooks/permission';
import { manageMenus, appMenus } from '@/router/app-menus';
import { cloneDeep } from 'lodash';

export default function useMenuTree() {
  const permission = usePermission();
  const comManageMenus = computed(() => {
    return manageMenus;
  });
  const comAppMenus = computed(() => {
    return appMenus;
  });
  function travel(_routes: RouteRecordRaw[], layer: number) {
    if (!_routes) return null;

    const collector: any = _routes.map((element) => {
      // no access
      if (!permission.accessRouter(element)) {
        return null;
      }

      // leaf node
      if (element.meta?.hideChildrenInMenu || !element.children) {
        element.children = [];
        return element;
      }

      // route filter hideInMenu true
      element.children = element.children.filter(
        (x) => x.meta?.hideInMenu !== true
      );

      // Associated child node
      const subItem = travel(element.children, layer + 1);

      if (subItem.length) {
        element.children = subItem;
        return element;
      }
      // the else logic
      if (layer > 1) {
        element.children = subItem;
        return element;
      }

      if (element.meta?.hideInMenu === false) {
        return element;
      }

      return null;
    });
    return collector.filter(Boolean);
  }
  const manageMenuTree = computed(() => {
    const copyRouter = cloneDeep(
      comManageMenus.value
    ) as RouteRecordNormalized[];
    copyRouter.sort((a: RouteRecordNormalized, b: RouteRecordNormalized) => {
      return (a.meta.order || 0) - (b.meta.order || 0);
    });

    return travel(copyRouter, 0);
  });
  const appMenuTree = computed(() => {
    const copyRouter = cloneDeep(comAppMenus.value) as RouteRecordNormalized[];
    copyRouter.sort((a: RouteRecordNormalized, b: RouteRecordNormalized) => {
      return (a.meta.order || 0) - (b.meta.order || 0);
    });

    return travel(copyRouter, 0);
  });

  return {
    manageMenuTree,
    appMenuTree,
  };
}
