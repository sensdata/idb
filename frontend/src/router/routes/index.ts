import manageRoutes from './manage';
import appRoutes from './app';

export { manageRoutes, appRoutes };
export const allRoutes = [...manageRoutes, ...appRoutes];
