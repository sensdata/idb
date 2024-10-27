import { ref } from 'vue';

export default function useVisible(initValue = false) {
  const visible = ref(initValue);
  const setVisible = (value: boolean) => {
    visible.value = value;
  };
  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };
  const toggle = () => {
    visible.value = !visible.value;
  };
  return {
    visible,
    show,
    hide,
    setVisible,
    toggle,
  };
}
