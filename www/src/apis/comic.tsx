const API_BASE = '/api';

export interface Comic {
  id: number;
  title: string;
  brief: string;
  icon: string;
  bg: string;
  user_prompt: string;
  status: 'pending' | 'completed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface ComicRole {
  id: number;
  comic_id: number;
  name: string;
  image_url: string;
  brief: string;
  voice: string;
  created_at: string;
  updated_at: string;
}

export interface ComicSection {
  id: number;
  comic_id: number;
  index: number;
  detail: string;
  status: 'pending' | 'completed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface ComicDetail extends Comic {
  roles: ComicRole[];
  sections: ComicSection[];
}

export interface ComicStoryboardDetail {
  id: number;
  storyboard_id: number;
  detail: string;
  role_id: number;
  tts_url: string;
  created_at: string;
  updated_at: string;
}

export interface ComicStoryboard {
  id: number;
  section_id: number;
  image_prompt: string;
  image_url: string;
  created_at: string;
  updated_at: string;
  details: ComicStoryboardDetail[];
}

export interface SectionContent {
  section: ComicSection;
  storyboards: ComicStoryboard[];
}

export interface GetComicsParams {
  page?: number;
  limit?: number;
  status?: 'pending' | 'completed' | 'failed';
}

export interface GetComicsResponse {
  code: number;
  message: string;
  data: {
    comics: Comic[];
    total: number;
    page: number;
    limit: number;
  };
}

export interface CreateComicParams {
  title: string;
  file: File;
  user_prompt: string;
}

export interface CreateComicResponse {
  code: number;
  message: string;
  data: Comic;
}

export interface GetComicResponse {
  code: number;
  message: string;
  data: ComicDetail;
}

export interface GetComicSectionsResponse {
  code: number;
  message: string;
  data: ComicSection[];
}

export interface CreateSectionParams {
  index: number;
  file: File;
}

export interface CreateSectionResponse {
  code: number;
  message: string;
  data: ComicSection;
}

export interface GetSectionContentResponse {
  code: number;
  message: string;
  data: SectionContent;
}

export interface GetStoryboardsResponse {
  code: number;
  message: string;
  data: ComicStoryboard[];
}

export async function getComics(params?: GetComicsParams): Promise<GetComicsResponse> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', params.page.toString());
  if (params?.limit) queryParams.append('limit', params.limit.toString());
  if (params?.status) queryParams.append('status', params.status);

  const url = `${API_BASE}/comic${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  const response = await fetch(url);
  return response.json();
}

export async function createComic(params: CreateComicParams): Promise<CreateComicResponse> {
  const formData = new FormData();
  formData.append('title', params.title);
  formData.append('file', params.file);
  formData.append('user_prompt', params.user_prompt);

  const response = await fetch(`${API_BASE}/comic`, {
    method: 'POST',
    body: formData,
  });
  return response.json();
}

export async function getComic(id: number): Promise<GetComicResponse> {
  const response = await fetch(`${API_BASE}/comic/${id}`);
  return response.json();
}

export async function getComicSections(id: number): Promise<GetComicSectionsResponse> {
  const response = await fetch(`${API_BASE}/comic/${id}/sections`);
  return response.json();
}

export async function createSection(id: number, params: CreateSectionParams): Promise<CreateSectionResponse> {
  const formData = new FormData();
  formData.append('index', params.index.toString());
  formData.append('file', params.file);

  const response = await fetch(`${API_BASE}/comic/${id}/section`, {
    method: 'POST',
    body: formData,
  });
  return response.json();
}

export async function getSectionContent(id: number, sectionId: number): Promise<GetSectionContentResponse> {
  const response = await fetch(`${API_BASE}/comic/${id}/section/${sectionId}/content`);
  return response.json();
}

export async function getStoryboards(id: number, sectionId: number): Promise<GetStoryboardsResponse> {
  const response = await fetch(`${API_BASE}/comic/${id}/section/${sectionId}/storyboards`);
  return response.json();
}

export async function getStoryboardImage(id: number, sectionId: number, storyboardId: number): Promise<Blob> {
  const response = await fetch(`${API_BASE}/comic/${id}/section/${sectionId}/storyboard/${storyboardId}/image`);
  return response.blob();
}
