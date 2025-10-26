export type ComicStatus = 'failed' | 'completed' | 'pending';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface ApiError {
  code: number;
  message: string;
  details?: string;
}

export interface Comic {
  id: string;
  icon_image_id: string;
  background_image_id: string;
  title: string;
  status: ComicStatus;
  created_at: string;
  updated_at: string;
}

export interface ComicDetail {
  id: string;
  icon_image_id: string;
  background_image_id: string;
  title: string;
  status: ComicStatus;
  roles: Role[];
  sections: Section[];
  created_at: string;
  updated_at: string;
}

export interface Role {
  name: string;
  brief: string;
  image_id: string;
  created_at: string;
  updated_at: string;
}

export interface Section {
  id: string;
  title: string;
  index: number;
  status: ComicStatus;
  created_at: string;
  updated_at: string;
}

export interface SectionDetail {
  id: string;
  title: string;
  index: number;
  status: ComicStatus;
  pages: Page[];
  created_at: string;
  updated_at: string;
}

export interface Page {
  id: string;
  created_at: string;
  updated_at: string;
  details: PageDetail[];
}

export interface PageDetail {
  id: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface GetComicsParams {
  page?: number;
  limit?: number;
}

export interface GetComicsData {
  comics: Comic[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateComicData {
  id: string;
}

export interface CreateSectionData {
  id: string;
  index: number;
}

export interface ImageUrlData {
  url: string;
}
