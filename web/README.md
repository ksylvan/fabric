# Fabric Web App

A user-friendly web interface for [Fabric](https://github.com/danielmiessler/Fabric) built with [Svelte 5](https://svelte.dev/), [SvelteKit 2](https://svelte.dev/docs/kit), [Vite 6](https://vite.dev/), [Tailwind CSS 4](https://tailwindcss.com/), and [Mdsvex](https://mdsvex.pngwn.io/).

![Fabric Web App Preview](../docs/images/svelte-preview.png)
_Alt: Screenshot of the Fabric web app dashboard showing pattern inputs and outputs._

## Table of Contents

- [Fabric Web App](#fabric-web-app)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Development](#development)
  - [Building for Production](#building-for-production)
  - [Available Scripts](#available-scripts)
  - [Obsidian Integration](#obsidian-integration)
  - [Contributing](#contributing)

## Prerequisites

- **Node.js** >= 18
- **Fabric** installed and available on your PATH (`fabric --version` to check)
- **Fabric server** running for backend API connectivity:

```bash
fabric --serve
```

This exposes Fabric's API at <http://localhost:8080>, which the web app uses for pattern execution, model listing, and chat streaming.

## Installation

From the `web/` directory:

**Using npm:**

```bash
cd web && npm install
```

**Or using pnpm:**

```bash
cd web && pnpm install
```

## Development

Make sure `fabric --serve` is running in a separate terminal, then start the dev server:

**Using npm:**

```bash
npm run dev
```

**Or using pnpm:**

```bash
pnpm run dev
```

Visit [http://localhost:5173](http://localhost:5173) (default port).

> [!TIP]
>
> Sync Svelte types if needed: `npx svelte-kit sync`

## Building for Production

```bash
npm run build
```

Preview the production build locally:

```bash
npm run preview
```

## Available Scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start the development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview the production build |
| `npm run check` | Run `svelte-check` for type checking |
| `npm run check:watch` | Run `svelte-check` in watch mode |
| `npm run test` | Run tests with Vitest |
| `npm run lint` | Check formatting (Prettier) and linting (ESLint) |
| `npm run format` | Apply Prettier formatting to all files |

## Obsidian Integration

Turn `web/src/lib/content/` into an [Obsidian](https://obsidian.md) vault for note-taking synced with Fabric patterns. It includes pre-configured `.obsidian/` and `templates/` folders.

### Quick Setup

1. Open Obsidian: File > Open folder as vault > Select `web/src/lib/content/`
2. To publish posts, move them to the posts directory (`web/src/lib/content/posts`).
3. Use Fabric patterns to generate content directly in Markdown files.

> [!TIP]
>
> When creating new posts, make sure to include a date (YYYY-MM-DD), description, tags (e.g., #ai #patterns), and aliases for SEO. Only a date is needed to display a note. Embed images `(![alt](path))`, link patterns `([[pattern-name]])`, or code blocks for reusable snippets — all in standard Markdown.

## Contributing

Refer to the [Contributing Guide](/docs/CONTRIBUTING.md) for details on how to improve this content.
