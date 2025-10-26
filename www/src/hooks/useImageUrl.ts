import { useState, useEffect } from 'react';
import { getImageUrl } from '@/apis';

interface UseImageUrlResult {
  url: string | null;
  loading: boolean;
  error: string | null;
}

const MAX_POLL_ATTEMPTS = 30;
const POLL_INTERVAL = 2000;

export function useImageUrl(imageId: string | undefined | null): UseImageUrlResult {
  const [url, setUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!imageId) {
      setUrl(null);
      setLoading(false);
      setError(null);
      return;
    }

    let attempts = 0;
    let timeoutId: NodeJS.Timeout;
    let isCancelled = false;

    const pollImageUrl = async () => {
      if (isCancelled) return;

      try {
        const response = await getImageUrl(imageId);
        
        if (response.code === 200 && response.data.url) {
          setUrl(response.data.url);
          setLoading(false);
          setError(null);
          return;
        }

        attempts++;
        
        if (attempts >= MAX_POLL_ATTEMPTS) {
          setError('Image loading timeout');
          setLoading(false);
          return;
        }

        timeoutId = setTimeout(pollImageUrl, POLL_INTERVAL);
      } catch (err) {
        if (!isCancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load image');
          setLoading(false);
        }
      }
    };

    setLoading(true);
    setError(null);
    pollImageUrl();

    return () => {
      isCancelled = true;
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  }, [imageId]);

  return { url, loading, error };
}
