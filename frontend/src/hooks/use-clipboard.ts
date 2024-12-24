import Clipboard from 'clipboard';

interface Options {
  appendToBody: boolean;
}

export function useClipboard(opts?: Options) {
  const appendToBody =
    opts?.appendToBody === undefined ? true : opts.appendToBody;
  return {
    copyText(text: string, container?: HTMLElement) {
      return new Promise((resolve, reject) => {
        const ua = navigator.userAgent;
        const isIOS = /iPad|iPhone|iPod/.test(ua);
        const isMobile =
          /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
            ua
          );
        if (isMobile && isIOS) {
          const textareaEl = document.createElement('textarea');
          textareaEl.value = String(text) || '';
          textareaEl.style.position = 'fixed';
          document.body.appendChild(textareaEl);
          textareaEl.focus();
          textareaEl.setSelectionRange(0, textareaEl.value.length);
          try {
            const success = document.execCommand('copy');
            if (success) {
              resolve({ text });
            } else {
              reject(new Error('copy failed'));
            }
          } catch (err) {
            reject(err);
          }
          document.body.removeChild(textareaEl);
        } else {
          const fakeEl = document.createElement('button');
          const clipboard = new Clipboard(fakeEl, {
            text: () => text,
            action: () => 'copy',
            container: container !== undefined ? container : document.body,
          });
          clipboard.on('success', (e) => {
            clipboard.destroy();
            resolve(e);
          });
          clipboard.on('error', (e) => {
            clipboard.destroy();
            reject(e);
          });
          if (appendToBody) document.body.appendChild(fakeEl);
          fakeEl.click();
          if (appendToBody) document.body.removeChild(fakeEl);
        }
      });
    },
  };
}
