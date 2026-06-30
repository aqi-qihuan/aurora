// Portfolio 作品集相关类型定义

export interface PortfolioItem {
  id: number
  name: string
  fullName?: string
  description?: string
  htmlUrl: string
  homepage?: string
  language?: string
  stargazersCount: number
  forksCount: number
  topics?: string[]
  repoUpdatedAt?: string
  cover?: string
  sort?: number
  isFeatured?: number
}
