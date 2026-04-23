import type { Paginated } from '@/types'

export function delay<T>(value: T, ms = 120): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms))
}

export function paginate<T>(
  items: T[],
  page = 1,
  pageSize = 20,
): Paginated<T> {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    page,
    pageSize,
    total: items.length,
  }
}
