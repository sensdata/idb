interface CookieOptions {
  expires?: number | Date;
  path?: string;
  domain?: string;
  secure?: boolean;
  sameSite?: 'strict' | 'lax' | 'none';
  httpOnly?: boolean;
  maxAge?: number;
}

export default class cookie {
  static get(name: string) {
    const cookies = document.cookie.split(';');
    for (const ck of cookies) {
      const [cookieName, cookieValue] = ck.trim().split('=');
      if (cookieName === name) {
        return decodeURIComponent(cookieValue);
      }
    }
    return undefined;
  }

  static set(name: string, value: string, options?: CookieOptions) {
    let cookieString = `${encodeURIComponent(name)}=${encodeURIComponent(
      value
    )}`;

    if (options) {
      if (options.expires) {
        if (typeof options.expires === 'number') {
          const d = new Date();
          d.setTime(d.getTime() + options.expires * 24 * 60 * 60 * 1000);
          cookieString += `;expires=${d.toUTCString()}`;
        } else {
          cookieString += `;expires=${options.expires.toUTCString()}`;
        }
      }
      if (options.maxAge) cookieString += `;max-age=${options.maxAge}`;
      if (options.path) cookieString += `;path=${options.path}`;
      if (options.domain) cookieString += `;domain=${options.domain}`;
      if (options.secure) cookieString += ';secure';
      if (options.httpOnly) cookieString += ';httponly';
      if (options.sameSite) cookieString += `;samesite=${options.sameSite}`;
    }

    document.cookie = cookieString;
    return value;
  }

  static remove(name: string) {
    const date = new Date();
    date.setTime(date.getTime() - 864e5);
    document.cookie = `${encodeURIComponent(
      name
    )}=;expires=${date.toUTCString()}; path=/`;
  }
}
