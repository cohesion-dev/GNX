const API_BASE = '/api';

export async function getTTSAudio(storyboardTtsId: number): Promise<Blob> {
  const response = await fetch(`${API_BASE}/tts/${storyboardTtsId}`);
  return response.blob();
}
