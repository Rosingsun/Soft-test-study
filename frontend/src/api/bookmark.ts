import { get, post, del } from './request'
import type { FavoriteFolder, CreateFolderReq, QuestionFavoriteResp } from '@/types/question'

export function createFolder(data: CreateFolderReq) {
  return post<FavoriteFolder>('/favorite-folders', data)
}

export function listFolders() {
  return get<FavoriteFolder[]>('/favorite-folders')
}

export function addFavorite(questionId: number, folderId?: number) {
  const query = folderId ? `?folder_id=${folderId}` : ''
  return post<{ favorited: boolean }>(`/questions/${questionId}/favorite${query}`)
}

export function removeFavorite(questionId: number) {
  return del<{ favorited: boolean }>(`/questions/${questionId}/favorite`)
}

export function checkFavorited(questionId: number) {
  return get<{ favorited: boolean }>(`/questions/${questionId}/favorited`)
}

export function listFavorites(folderId?: number) {
  const query = folderId ? `?folder_id=${folderId}` : ''
  return get<QuestionFavoriteResp[]>(`/favorites${query}`)
}
