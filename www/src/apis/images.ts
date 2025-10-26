import type { ApiResponse, ImageUrlData } from './types';
import { API_BASE } from './config';

export async function getImageUrl(imageId: string): Promise<ApiResponse<ImageUrlData>> {
  const response = await fetch(`${API_BASE}/images/${imageId}/url/`);
  return response.json();
}
