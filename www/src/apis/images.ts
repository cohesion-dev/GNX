import type { ApiResponse, ImageUrlData } from './types';

const API_BASE = '/api';

export async function getImageUrl(imageId: string): Promise<ApiResponse<ImageUrlData>> {
  const response = await fetch(`${API_BASE}/images/${imageId}/url`);
  return response.json();
}
