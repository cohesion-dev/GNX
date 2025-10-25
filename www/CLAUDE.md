# GNX WWW Frontend

This is the frontend application for GNX, a comic creation platform built with Next.js.

## Tech Stack

- **Framework**: Next.js 16.0.0 (App Router)
- **React**: 19.2.0
- **State Management**: MobX 6.15.0 with mobx-react-lite
- **Styling**: Tailwind CSS 4.1.16
- **TypeScript**: 5.9.3

## Project Structure

```
www/
├── src/
│   ├── app/              # Next.js App Router pages
│   │   ├── page.tsx     # Homepage with redirect logic
│   │   └── comic/       # Comic-related pages
│   │       ├── page.tsx           # Comic list page
│   │       ├── add/page.tsx       # Create new comic page
│   │       ├── detail/[id]/       # Comic detail pages
│   │       └── read/[id]/         # Comic reading page
│   ├── apis/            # API client functions
│   │   └── comic.tsx    # Comic-related API calls
│   ├── components/      # Reusable React components
│   └── styles/          # Global styles
├── package.json
├── tsconfig.json
└── next.config.js
```

## Key Features

### Homepage Logic (page.tsx)
- Fetches comic list on load
- If no comics exist (length = 0), redirects to `/comic/add`
- If comics exist (length > 0), redirects to `/comic/` (comic list page)
- Handles errors gracefully by defaulting to comic list page

### API Integration
The `src/apis/comic.tsx` file provides functions for:
- `getComics()`: Fetch comic list with pagination
- `createComic()`: Create new comic
- `getComic()`: Get comic details
- `getComicSections()`: Get comic sections
- `createSection()`: Create new section
- `getSectionContent()`: Get section content
- `getStoryboards()`: Get storyboards for a section

### Data Models
- **Comic**: Base comic information (id, title, brief, icon, bg, status, etc.)
- **ComicRole**: Character/role in a comic
- **ComicSection**: Story sections within a comic
- **ComicStoryboard**: Visual storyboards for sections

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

## Conventions

- Use `'use client'` directive for client components
- Follow Next.js App Router patterns
- Use dynamic imports for performance optimization
- Use MobX observer for reactive components
- Use Tailwind CSS for styling
- API base path: `/api`
- Path alias: `@/*` maps to `src/*`

## Notes

- The app uses responsive design with separate PC and Mobile components
- Components are dynamically loaded with `{ ssr: false }` for client-side rendering
- All API responses follow a consistent format: `{ code, message, data }`
