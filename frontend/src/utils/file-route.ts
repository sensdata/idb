/**
 * 文件管理路由工具函数
 */

/**
 * 根据文件路径生成路由对象
 * @param filePath 文件路径，如 '/home/user/documents'
 * @param query 额外的查询参数
 * @returns Vue Router 路由对象
 */
export function createFileRoute(filePath: string, query?: Record<string, any>) {
  const finalQuery = { ...query };

  // 将路径作为查询参数传递
  if (filePath && filePath !== '/') {
    finalQuery.path = filePath;
  }

  return {
    name: 'file',
    query: finalQuery,
  };
}

/**
 * 根据文件路径和分页信息生成路由对象
 * @param filePath 文件路径
 * @param pagination 分页信息
 * @param query 额外的查询参数
 * @returns Vue Router 路由对象
 */
export function createFileRouteWithPagination(
  filePath: string,
  pagination?: {
    page?: number;
    pageSize?: number;
  },
  query?: Record<string, any>
) {
  const finalQuery = { ...query };

  // 将路径作为查询参数传递
  if (filePath && filePath !== '/') {
    finalQuery.path = filePath;
  }

  // 只有当页码不是1时才添加到查询参数中
  if (pagination?.page && pagination.page > 1) {
    finalQuery.page = pagination.page.toString();
  }

  // 只有当页面大小不是默认值(20)时才添加到查询参数中
  if (pagination?.pageSize && pagination.pageSize !== 20) {
    finalQuery.pageSize = pagination.pageSize.toString();
  }

  return {
    name: 'file',
    query: finalQuery,
  };
}

/**
 * 从路由查询参数中解析文件路径
 * @param routeQuery 路由查询参数对象
 * @returns 解析后的文件路径
 */
export function parseFilePathFromRoute(routeQuery: any): string {
  const filePath = routeQuery.path;

  if (!filePath || filePath === '') {
    return '/';
  }

  if (typeof filePath === 'string') {
    // 确保路径以 / 开头
    return filePath.startsWith('/') ? filePath : `/${filePath}`;
  }

  // 处理数组情况（虽然在查询参数中不太可能，但保持兼容性）
  if (Array.isArray(filePath)) {
    const joinedPath = filePath.join('/');
    return joinedPath.startsWith('/') ? joinedPath : `/${joinedPath}`;
  }

  return '/';
}

/**
 * 从路由查询参数中解析分页信息
 * @param routeQuery 路由查询参数对象
 * @returns 解析后的分页信息
 */
export function parsePaginationFromRoute(routeQuery: any): {
  page: number;
  pageSize: number;
} {
  const page = parseInt(routeQuery.page, 10) || 1;
  const pageSize = parseInt(routeQuery.pageSize, 10) || 20;

  return {
    page: Math.max(1, page),
    pageSize: Math.max(10, Math.min(100, pageSize)),
  };
}

/**
 * 生成文件管理页面的完整URL
 * @param filePath 文件路径
 * @param options 选项
 * @returns 完整的URL字符串
 */
export function generateFileManagementUrl(
  filePath: string,
  options?: {
    hostId?: number;
    page?: number;
    pageSize?: number;
  }
): string {
  const queryParams = new URLSearchParams();

  // 添加路径参数
  if (filePath && filePath !== '/') {
    queryParams.set('path', filePath);
  }

  if (options?.hostId) {
    queryParams.set('id', options.hostId.toString());
  }

  if (options?.page && options.page > 1) {
    queryParams.set('page', options.page.toString());
  }

  if (options?.pageSize && options.pageSize !== 20) {
    queryParams.set('pageSize', options.pageSize.toString());
  }

  const queryString = queryParams.toString();
  const basePath = '/app/file';

  if (queryString) {
    return `${basePath}?${queryString}`;
  }

  return basePath;
}
