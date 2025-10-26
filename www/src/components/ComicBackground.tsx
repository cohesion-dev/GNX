'use client'

import { useImageUrl } from '@/hooks/useImageUrl'

interface ComicBackgroundProps {
  imageId: string | undefined | null;
  alt?: string;
  className?: string;
}

export default function ComicBackground({ imageId, alt = 'Comic background', className = '' }: ComicBackgroundProps) {
  const { url, loading, error } = useImageUrl(imageId);

  if (error) {
    return (
      <div className={`flex items-center justify-center bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 ${className}`}>
        <span className="text-4xl">📚</span>
      </div>
    );
  }

  if (loading || !url) {
    return (
      <div className={`flex items-center justify-center bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 ${className}`}>
        <svg
          className="w-8 h-8 animate-spin"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      </div>
    );
  }

  return (
    <img 
      src={url} 
      alt={alt}
      className={className}
    />
  );
}
