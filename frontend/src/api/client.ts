import axios from 'axios'
import type {
  SearchRespData,
  GetComicRespData,
  ChapterItem,
  GetChapterRespData,
  DownloadTask,
  DownloadedComic
} from '../types'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 搜索漫画
export async function search(keyword: string, page: number = 1): Promise<SearchRespData> {
  const { data } = await api.get('/search', {
    params: { q: keyword, page }
  })
  if (!data.success) {
    throw new Error(data.message || '搜索失败')
  }
  return data.data
}

// 获取漫画详情
export async function getComic(pathWord: string): Promise<GetComicRespData> {
  const { data } = await api.get(`/comic/${pathWord}`)
  if (!data.success) {
    throw new Error(data.message || '获取漫画失败')
  }
  return data.data
}

// 获取分组章节
export async function getGroupChapters(
  comicPathWord: string,
  groupPathWord: string
): Promise<ChapterItem[]> {
  const { data } = await api.get(`/comic/${comicPathWord}/group/${groupPathWord}/chapters`)
  if (!data.success) {
    throw new Error(data.message || '获取章节失败')
  }
  return data.data
}

// 获取章节图片
export async function getChapterImages(
  comicPathWord: string,
  chapterUUID: string
): Promise<GetChapterRespData> {
  const { data } = await api.get(`/comic/${comicPathWord}/chapter/${chapterUUID}`)
  if (!data.success) {
    throw new Error(data.message || '获取章节图片失败')
  }
  return data.data
}

// 开始下载
export async function startDownload(
  comicName: string,
  comicPathWord: string,
  chapterUUID: string,
  chapterTitle: string
): Promise<string> {
  const { data } = await api.post('/download', {
    comic_name: comicName,
    comic_path_word: comicPathWord,
    chapter_uuid: chapterUUID,
    chapter_title: chapterTitle
  })
  if (!data.success) {
    throw new Error(data.message || '创建下载任务失败')
  }
  return data.task_id
}

// 获取下载任务列表
export async function getTasks(): Promise<DownloadTask[]> {
  const { data } = await api.get('/tasks')
  return data
}

// 获取已下载漫画列表
export async function getDownloadedComics(): Promise<DownloadedComic[]> {
  const { data } = await api.get('/downloaded')
  if (!data.success) {
    throw new Error(data.message || '获取已下载漫画失败')
  }
  return data.data
}

// 测试 API
export async function ping(): Promise<{ status: number; ok: boolean; body_preview: string }> {
  const { data } = await api.get('/ping')
  return data
}
