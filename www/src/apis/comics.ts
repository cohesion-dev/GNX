import type {
  ApiResponse,
  GetComicsParams,
  GetComicsData,
  CreateComicData,
  ComicDetail,
} from './types';
import { API_BASE } from './config';

export async function getComics(
  params?: GetComicsParams
): Promise<ApiResponse<GetComicsData>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', params.page.toString());
  if (params?.limit) queryParams.append('limit', params.limit.toString());

  const url = `${API_BASE}/comics/${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  const response = await fetch(url);
  return response.json();
}

export async function createComic(
  title: string,
  userPrompt: string,
  file: File
): Promise<ApiResponse<CreateComicData>> {
  const formData = new FormData();
  formData.append('title', title);
  formData.append('user_prompt', userPrompt);
  formData.append('file', file);

  const response = await fetch(`${API_BASE}/comics/`, {
    method: 'POST',
    body: formData,
  });
  return response.json();
}

export async function getComic(comicId: string): Promise<ApiResponse<ComicDetail>> {
  const response = await fetch(`${API_BASE}/comics/${comicId}/`);
  return response.json();
}
