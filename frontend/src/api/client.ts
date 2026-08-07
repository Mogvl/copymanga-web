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
  chapterTitle: string,
  groupTitle: string,
  order: number,
  imageFormat?: string
): Promise<string> {
  const { data } = await api.post('/download', {
    comic_name: comicName,
    comic_path_word: comicPathWord,
    chapter_uuid: chapterUUID,
    chapter_title: chapterTitle,
    group_title: groupTitle,
    order: order,
    image_format: imageFormat || 'webp'
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

// 下载并触发浏览器保存 PDF blob
function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

// 解析可能为错误 JSON 的 blob
async function asError(blob: Blob): Promise<string> {
  try {
    const text = await blob.text()
    const parsed = JSON.parse(text)
    return parsed.message || parsed.error || text
  } catch {
    return '导出失败'
  }
}

// 导出单章 PDF
export async function exportChapterPdf(
  params: {
    comicName: string
    comicPathWord: string
    chapterUUID: string
    groupTitle: string
    order: number
    chapterTitle: string
    useLocalOnly?: boolean
  },
  filename: string
): Promise<void> {
  const { data } = await api.post('/export/chapter-pdf', {
    comic_name: params.comicName,
    comic_path_word: params.comicPathWord,
    chapter_uuid: params.chapterUUID,
    group_title: params.groupTitle,
    order: params.order,
    chapter_title: params.chapterTitle,
    use_local_only: params.useLocalOnly || false
  }, { responseType: 'blob' })
  if (data instanceof Blob && data.type === 'application/json') {
    throw new Error(await asError(data))
  }
  downloadBlob(data as Blob, filename)
}

// 导出整本 PDF
export async function exportComicPdf(
  params: { comicName: string; comicPathWord: string; useLocalOnly?: boolean },
  filename: string
): Promise<void> {
  const { data } = await api.post('/export/comic-pdf', {
    comic_name: params.comicName,
    comic_path_word: params.comicPathWord,
    use_local_only: params.useLocalOnly || false
  }, { responseType: 'blob' })
  if (data instanceof Blob && data.type === 'application/json') {
    throw new Error(await asError(data))
  }
  downloadBlob(data as Blob, filename)
}

// 测试 API
export async function ping(): Promise<{ status: number; ok: boolean; body_preview: string }> {
  const { data } = await api.get('/ping')
  return data
}
