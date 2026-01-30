export interface Package {
  id: string
  name: string
  description: string
  dependencies: string[]
  estimatedTime: number // minutes
  category: Category
  size: string
}

export type Category =
  | 'Foundation'
  | 'ML Framework'
  | 'LLM Runtime'
  | 'LLM Models'
  | 'Image Generation'
  | 'Video Generation'
  | 'Audio/Voice'
  | 'Vision Models'
  | 'Application'
  | 'Tools'

export interface InstallProgress {
  packageId: string
  status: 'pending' | 'installing' | 'completed' | 'failed'
  progress: number
  message: string
  startTime?: number
  endTime?: number
}

export interface Server {
  id: string
  name: string
  host: string
  port: number
  username: string
  status: 'connected' | 'disconnected' | 'error'
}

export const CategoryEmojis: Record<Category, string> = {
  'Foundation': '🏗️',
  'ML Framework': '🤖',
  'LLM Runtime': '🔮',
  'LLM Models': '⭐',
  'Image Generation': '🖼️',
  'Video Generation': '🎬',
  'Audio/Voice': '🎙️',
  'Vision Models': '👁️',
  'Application': '🎯',
  'Tools': '🔧',
}

export const CategoryColors: Record<Category, string> = {
  'Foundation': 'electric',
  'ML Framework': 'neon',
  'LLM Runtime': 'sakura',
  'LLM Models': 'sunset',
  'Image Generation': 'electric',
  'Video Generation': 'mint',
  'Audio/Voice': 'neon',
  'Vision Models': 'sakura',
  'Application': 'electric',
  'Tools': 'mint',
}
