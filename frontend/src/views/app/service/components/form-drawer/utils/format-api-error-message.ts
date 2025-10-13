export function formatApiErrorMessage(error: any, fallback: string): string {
  try {
    const responseData = error?.response?.data;
    const base =
      responseData?.message ||
      error?.message ||
      (typeof error === 'string' ? error : '') ||
      fallback;

    const data = responseData?.data;
    let detail = '';
    if (typeof data === 'string') {
      detail = data.trim();
    } else if (data && typeof data === 'object') {
      try {
        const s = JSON.stringify(data, null, 2);
        if (s !== '{}' && s !== '[]' && s !== 'null') detail = s;
      } catch {
        /* ignore */
      }
    } else if (data != null) {
      detail = String(data);
    }

    return detail ? `${base}\n\n${detail}` : base;
  } catch {
    return error?.message || fallback;
  }
}
