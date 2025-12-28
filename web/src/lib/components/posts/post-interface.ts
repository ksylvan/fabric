import type { SvelteComponent } from 'svelte';
import type { Frontmatter } from '$lib/utils/markdown';

export type PostMetadata = Frontmatter;

/** Post summary for listing pages (no content) */
export interface PostSummary {
    /** URL-friendly identifier for the post */
    slug: string;
    /** Post metadata from frontmatter */
    metadata: PostMetadata;
}

/** Full post with content */
export interface Post extends PostSummary {
    /** Compiled Svelte component or HTML string */
    content: string | typeof SvelteComponent;
}
