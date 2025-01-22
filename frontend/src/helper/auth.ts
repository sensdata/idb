import cookie from './cookie';

const TOKEN_KEY = 'idb-token';

const isLogin = () => {
  return !!cookie.get(TOKEN_KEY);
};

const getToken = () => {
  return cookie.get(TOKEN_KEY);
};

const setToken = (token: string) => {
  cookie.set(TOKEN_KEY, token, { expires: 3, path: '/' });
};

const clearToken = () => {
  cookie.remove(TOKEN_KEY);
};

export { isLogin, getToken, setToken, clearToken };
