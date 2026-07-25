export interface AuthorInfo {
  name: string
  alias?: string
  path_word: string
}

export interface ComicInSearch {
  name: string
  alias?: string
  path_word: string
  cover: string
  ban: number
  author: AuthorInfo[]
  popular: number
}

export interface SearchRespData {
  list: ComicInSearch[]
  total: number
  limit: number
  offset: number
}

export interface LabeledValue {
  value: number
  display: string
}

export interface ThemeInfo {
  name: string
  path_word: string
}

export interface Group {
  path_word: string
  count: number
  name: string
}

export interface ComicDetail {
  uuid: string
  name: string
  path_word: string
  author: AuthorInfo[]
  cover: string
  brief: string
  status: LabeledValue
  theme: ThemeInfo[]
  datetime_updated: string
}

export interface GetComicRespData {
  is_banned: boolean
  comic: ComicDetail
  popular: number
  groups: Record<string, Group>
}

export interface ChapterItem {
  index: number
  uuid: string
  count: number
  ordered: number
  size: number
  name: string
  comic_path_word: string
  group_path_word: string
  datetime_created: string
}

export interface ContentURL {
  url: string
}

export interface ChapterContent {
  uuid: string
  name: string
  comic_path_word: string
  group_path_word: string
  contents: ContentURL[]
  words: number[]
}

export interface GetChapterRespData {
  chapter: ChapterContent
}

export interface DownloadTask {
  id: string
  comic_name: string
  comic_path_word: string
  chapter_uuid: string
  chapter_title: string
  status: 'pending' | 'downloading' | 'completed' | 'failed'
  progress: string
  total_pages: number
  downloaded_pages: number
  created_at: string
}

export interface DownloadedComic {
  name: string
  path_word: string
  chapter_count: number
  total_pages: number
}
