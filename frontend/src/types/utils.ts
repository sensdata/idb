import { AllowedComponentProps, Component, VNodeProps } from 'vue';

export type ComponentProps<C extends Component> = C extends new (...args: any) => {
  $props: infer P;
}
  ? Omit<P, keyof VNodeProps | keyof AllowedComponentProps>
  : never;
