# GNX WWW Frontend

This is the frontend application for GNX, a comic creation platform built with Next.js.

## Tech Stack

- **Framework**: Next.js 16.0.0 (App Router)
- **React**: 19.2.0
- **State Management**: MobX 6.15.0 with mobx-react-lite
- **Styling**: Tailwind CSS 4.1.16 with PostCSS
- **TypeScript**: 5.9.3

## Project Structure

```
www/
├── src/
│   ├── app/                        # Next.js App Router pages
│   │   ├── page.tsx                # Homepage with redirect logic
│   │   ├── layout.tsx              # Root layout component
│   │   └── comic/                  # Comic-related pages
│   │       ├── page.tsx            # Comic list page
│   │       ├── mobile.tsx          # Mobile view for comic list
│   │       ├── pc.tsx              # Desktop view for comic list
│   │       ├── add/                # Create new comic
│   │       │   ├── page.tsx
│   │       │   ├── mobile.tsx
│   │       │   └── pc.tsx
│   │       ├── detail/[id]/        # Comic detail pages
│   │       │   ├── page.tsx
│   │       │   ├── mobile.tsx
│   │       │   ├── pc.tsx
│   │       │   └── section/add/    # Add comic section
│   │       └── read/[id]/          # Comic reading page
│   │           ├── page.tsx
│   │           ├── mobile.tsx
│   │           └── pc.tsx
│   ├── apis/                       # API client modules
│   │   ├── comics.ts               # Comic CRUD operations
│   │   ├── sections.ts             # Section operations
│   │   ├── images.ts               # Image operations
│   │   ├── tts.ts                  # Text-to-speech API
│   │   ├── config.ts               # API configuration
│   │   ├── types.ts                # TypeScript type definitions
│   │   └── index.ts                # API exports
│   ├── stores/                     # MobX state management
│   │   ├── ComicReadStore.ts       # Comic reading state
│   │   ├── PageManager.ts          # Page navigation state
│   │   ├── AudioPlayer.ts          # Audio playback state
│   │   └── index.ts                # Store exports
│   ├── components/                 # Reusable React components
│   │   ├── ComicBackground.tsx     # Comic background component
│   │   ├── ComicIcon.tsx           # Comic icon component
│   │   └── layout.tsx              # Layout component
│   ├── hooks/                      # Custom React hooks
│   │   ├── useImageUrl.ts          # Image URL resolution hook
│   │   └── usePolling.ts           # Polling hook
│   └── styles/                     # Global styles
│       └── globals.css             # Global CSS styles
├── public/                         # Static assets
├── package.json
├── tsconfig.json
├── next.config.js
├── postcss.config.js
└── .gitignore
```

## Key Features

### Homepage Logic (app/page.tsx)
- Fetches comic list on load
- If no comics exist (length = 0), redirects to `/comic/add`
- If comics exist (length > 0), redirects to `/comic/` (comic list page)
- Handles errors gracefully by defaulting to comic list page

### API Integration
The `src/apis/` directory provides modular API client functions:
- **comics.ts**: `getComics()`, `createComic()`, `getComic()`
- **sections.ts**: `getSections()`, `createSection()`, `getSectionDetail()`
- **images.ts**: `getImageUrl()` for image URL resolution
- **tts.ts**: Text-to-speech functionality
- **types.ts**: Shared TypeScript type definitions

### Data Models (src/apis/types.ts)
- **Comic**: Base comic information (id, title, brief, icon_image_id, background_image_id, status, timestamps)
- **ComicDetail**: Extended comic with roles and sections
- **Role**: Character/role in a comic
- **Section**: Story sections within a comic
- **SectionDetail**: Section with page details
- **Page**: Comic pages with details
- **PageDetail**: Page content details
- **ComicStatus**: `'pending' | 'completed' | 'failed'`

### State Management (MobX Stores)
- **ComicReadStore**: Manages comic reading state, page navigation, and audio playback
- **PageManager**: Handles page navigation and transitions
- **AudioPlayer**: Controls audio playback for text-to-speech

### Responsive Design
- Separate mobile and desktop components for each page
- Dynamic component loading with `dynamic` from `next/dynamic`
- Client-side rendering for responsive components using `{ ssr: false }`

## Development

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Run linter
npm run lint
```

## Configuration

### Next.js (next.config.js)
- `reactStrictMode: false`
- `trailingSlash: true` - adds trailing slashes to URLs
- `images.unoptimized: true` - disables image optimization
- Environment variable: `NEXT_PUBLIC_API_BASE_URL`
- Empty rewrites array

### TypeScript (tsconfig.json)
- Target: ES2017
- Strict mode enabled
- Path alias: `@/*` maps to `src/*`
- JSX: `react-jsx`
- Module resolution: `bundler`

### PostCSS (postcss.config.js)
- Uses `@tailwindcss/postcss` plugin
- Autoprefixer enabled

## Conventions

- Use `'use client'` directive for client components
- Follow Next.js App Router patterns
- Use dynamic imports for performance optimization: `dynamic(() => import('./component'), { ssr: false })`
- Use MobX observer for reactive components
- Use Tailwind CSS for styling
- Separate PC and Mobile components for responsive design
- API base URL: configured via `NEXT_PUBLIC_API_BASE_URL` environment variable
- Path alias: `@/*` maps to `src/*`
- All API responses follow consistent format: `ApiResponse<T>` with `{ code, message, data }`

## Important Notes

- The app uses responsive design with separate PC and Mobile components
- Components are dynamically loaded with `{ ssr: false }` for client-side rendering
- All API calls use centralized configuration from `src/apis/config.ts`
- TypeScript types are centralized in `src/apis/types.ts`
- State management uses MobX with multiple specialized stores
- Image URLs are resolved through the `useImageUrl` hook
- Audio playback is managed by the AudioPlayer store
