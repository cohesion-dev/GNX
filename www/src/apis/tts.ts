import { API_BASE } from './config';

export async function getTTSAudio(ttsId: string): Promise<Blob> {
  const response = await fetch(`${API_BASE}/tts/${ttsId}/`);
  return response.blob();
}
