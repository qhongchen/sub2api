import { computed, nextTick, onBeforeUnmount, reactive, ref, shallowRef, type Ref } from 'vue'

type FetchOptions = {
  signal?: AbortSignal
}

type PageState = {
  page: number
  page_size: number
  total: number
}

type LoadMode = 'replace' | 'append' | 'silent'

export type InfinitePagedResponse<T, Extra = unknown> = {
  items?: T[]
  total?: number
  page?: number
  page_size?: number
} & Extra

type UseInfinitePagedListOptions<T, Response extends InfinitePagedResponse<T>> = {
  pageSize?: number
  sentinelRef: Ref<HTMLElement | null>
  fetchPage: (pagination: PageState, options: FetchOptions) => Promise<Response>
  onSuccess?: (response: Response, mode: LoadMode) => void
  onError?: (error: unknown, mode: LoadMode) => void
  isCanceled?: (error: unknown) => boolean
  canLoadMore?: () => boolean
  rootMargin?: string
  threshold?: number
}

export function useInfinitePagedList<T, Response extends InfinitePagedResponse<T>>(
  options: UseInfinitePagedListOptions<T, Response>
) {
  const items = shallowRef<T[]>([])
  const loading = ref(false)
  const loadingMore = ref(false)
  const refreshingSilently = ref(false)
  const pagination = reactive<PageState>({
    page: 1,
    page_size: options.pageSize ?? 20,
    total: 0
  })

  const hasMore = computed(() => items.value.length < pagination.total)

  let abortController: AbortController | null = null
  let loadMoreObserver: IntersectionObserver | null = null

  const canStartLoadMore = () => {
    return (
      hasMore.value &&
      !loading.value &&
      !loadingMore.value &&
      !refreshingSilently.value &&
      (options.canLoadMore?.() ?? true)
    )
  }

  const isSentinelVisible = () => {
    const el = options.sentinelRef.value
    if (!el) return false
    const rect = el.getBoundingClientRect()
    return rect.top <= window.innerHeight + 240 && rect.bottom >= -240
  }

  const scheduleLoadMoreIfNeeded = () => {
    void nextTick(() => {
      if (canStartLoadMore() && isSentinelVisible()) {
        void loadNextPage()
      }
    })
  }

  const load = async (mode: LoadMode = 'replace'): Promise<boolean> => {
    abortController?.abort()
    const controller = new AbortController()
    abortController = controller

    if (mode === 'append') {
      loadingMore.value = true
    } else if (mode === 'silent') {
      refreshingSilently.value = true
    } else {
      loading.value = true
      loadingMore.value = false
    }

    try {
      const response = await options.fetchPage(pagination, { signal: controller.signal })
      if (controller.signal.aborted) return false

      const nextItems = response.items || []
      items.value = mode === 'append' ? [...items.value, ...nextItems] : nextItems
      pagination.total = response.total || 0
      pagination.page = response.page || pagination.page
      pagination.page_size = response.page_size || pagination.page_size
      options.onSuccess?.(response, mode)
      scheduleLoadMoreIfNeeded()
      return true
    } catch (error) {
      if (options.isCanceled?.(error)) return false
      options.onError?.(error, mode)
      return false
    } finally {
      if (abortController === controller) {
        abortController = null
        loading.value = false
        loadingMore.value = false
        refreshingSilently.value = false
      }
    }
  }

  const reset = () => {
    abortController?.abort()
    pagination.page = 1
    pagination.total = 0
    items.value = []
    loadingMore.value = false
    refreshingSilently.value = false
    return load('replace')
  }

  const refreshFirstPageSilently = () => {
    abortController?.abort()
    pagination.page = 1
    loadingMore.value = false
    return load('silent')
  }

  async function loadNextPage(): Promise<boolean> {
    if (!canStartLoadMore()) return false
    const previousPage = pagination.page
    pagination.page = previousPage + 1
    const ok = await load('append')
    if (!ok && pagination.page === previousPage + 1) {
      pagination.page = previousPage
    }
    return ok
  }

  const stopObserver = () => {
    loadMoreObserver?.disconnect()
    loadMoreObserver = null
  }

  const startObserver = () => {
    stopObserver()
    if (typeof IntersectionObserver === 'undefined') return

    void nextTick(() => {
      const target = options.sentinelRef.value
      if (!target) return

      loadMoreObserver = new IntersectionObserver(
        (entries) => {
          if (entries.some((entry) => entry.isIntersecting)) {
            void loadNextPage()
          }
        },
        {
          root: null,
          rootMargin: options.rootMargin ?? '240px 0px',
          threshold: options.threshold ?? 0.01
        }
      )
      loadMoreObserver.observe(target)
    })
  }

  onBeforeUnmount(() => {
    abortController?.abort()
    stopObserver()
  })

  return {
    items,
    loading,
    loadingMore,
    refreshingSilently,
    pagination,
    hasMore,
    load,
    reset,
    refreshFirstPageSilently,
    loadNextPage,
    startObserver,
    stopObserver,
    scheduleLoadMoreIfNeeded
  }
}
