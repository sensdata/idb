import { createPinia } from 'pinia';
import useAppStore from './modules/app';
import useUserStore from './modules/user';
import useTabBarStore from './modules/tab-bar';
import useHostStore from './modules/host';

const pinia = createPinia();

export { useAppStore, useUserStore, useTabBarStore, useHostStore };
export default pinia;
