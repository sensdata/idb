import { computed, reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ApiListParams } from '@/types/global';

interface UrlSyncOptions {
  pageSize: number;
  urlParamNames?: {
    page?: string;
    pageSize?: string;
  };
}

export default function useUrlPaginationSync(options: UrlSyncOptions) {
  const route = useRoute();
  const router = useRouter();

  // URL同步配置
  const urlParamNames = computed(() => ({
    page: options.urlParamNames?.page || 'page',
    pageSize: options.urlParamNames?.pageSize || 'pageSize',
  }));

  // 从URL初始化分页参数
  const initFromUrl = () => {
    const pageParamName = urlParamNames.value.page;
    const pageSizeParamName = urlParamNames.value.pageSize;

    const pageFromUrl = parseInt(route.query[pageParamName] as string, 10) || 1;
    const pageSizeFromUrl =
      parseInt(route.query[pageSizeParamName] as string, 10) ||
      options.pageSize;

    return { page: pageFromUrl, pageSize: pageSizeFromUrl };
  };

  // 初始化分页参数
  const initialPagination = initFromUrl();

  const pagination = reactive({
    current: initialPagination.page,
    pageSize: initialPagination.pageSize,
  });

  // 参数对象
  const params = reactive<ApiListParams>({
    page: initialPagination.page,
    page_size: initialPagination.pageSize,
  });

  // URL同步功能
  const updateUrl = (page?: number, pageSize?: number) => {
    const query = { ...route.query };

    if (page !== undefined && page > 1) {
      query[urlParamNames.value.page] = page.toString();
    } else {
      delete query[urlParamNames.value.page];
    }

    if (pageSize !== undefined && pageSize !== options.pageSize) {
      query[urlParamNames.value.pageSize] = pageSize.toString();
    } else {
      delete query[urlParamNames.value.pageSize];
    }

    const newRoute = {
      name: route.name,
      params: route.params,
      query,
    };

    // 保持路径参数，只更新查询参数
    router.replace(newRoute);
  };

  // 更新分页参数
  const updatePagination = (page?: number, pageSize?: number) => {
    if (page !== undefined) {
      pagination.current = page;
      params.page = page;
    }

    if (pageSize !== undefined) {
      pagination.pageSize = pageSize;
      params.page_size = pageSize;

      // 切换每页条数时，重置到第一页
      if (page === undefined) {
        pagination.current = 1;
        params.page = 1;
      }
    }

    // 更新URL
    updateUrl(pagination.current, pagination.pageSize);

    return { ...params };
  };

  return {
    pagination,
    params,
    initFromUrl,
    updateUrl,
    updatePagination,
  };
}
