export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  next_page?: number
}

export interface ApiError {
  error: string
  code?: number
}
