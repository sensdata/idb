/* eslint-disable */
/* prettier-ignore */
// @ts-nocheck
import '@vue/runtime-core'

export {}

declare module '@vue/runtime-core' {
  export interface GlobalComponents {
    IdbTable: typeof import('@/components/idb-table/index.vue')["default"]
    IdbDropdownOperation: typeof import('@/components/idb-dropdown-operation/index.vue')["default"]
    FixedFooterBar: typeof import('@/components/fixed-form-bar/index.vue')["default"]
    Breadcrumb: typeof import('@/components/breadcrumb/index.vue')["default"]
  }
}
