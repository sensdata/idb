import type { RouteRecordNormalized } from 'vue-router';

const manageModules = import.meta.glob('./manage/*.ts', { eager: true });
const appModules = import.meta.glob('./app/*.ts', { eager: true });

function formatModules(_modules: any, result: RouteRecordNormalized[]) {
  Object.keys(_modules).forEach((key) => {
    const defaultModule = _modules[key].default;
    if (!defaultModule) return;
    const moduleList = Array.isArray(defaultModule)
      ? [...defaultModule]
      : [defaultModule];
    result.push(...moduleList);
  });
  return result;
}

export const manageRoutes = formatModules(manageModules, []);
export const appRoutes = formatModules(appModules, []);
export const allRoutes = [...manageRoutes, ...appRoutes];
