import type { ApiResponse, CreateSectionData, SectionDetail } from './types';

const API_BASE = '/api';

export async function createSection(
  comicId: string,
  title: string,
  content: string
): Promise<ApiResponse<CreateSectionData>> {
  const formData = new FormData();
  formData.append('title', title);
  formData.append('content', content);

  const response = await fetch(`${API_BASE}/comics/${comicId}/sections/`, {
    method: 'POST',
    body: formData,
  });
  return response.json();
}

export async function getSection(
  comicId: string,
  sectionId: string
): Promise<ApiResponse<SectionDetail>> {
  const response = await fetch(`${API_BASE}/comics/${comicId}/sections/${sectionId}/`);
  return response.json();
}
