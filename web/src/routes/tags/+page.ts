import type { PageLoad } from './$types';
import type { Frontmatter } from '$lib/utils/markdown';

interface TagPost {
	slug: string | undefined;
	metadata: Frontmatter;
}

export const load: PageLoad = async () => {
	const postFiles = import.meta.glob('/src/lib/content/posts/*.{md,svx}', { eager: true });

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const posts: TagPost[] = Object.entries(postFiles).map(([path, post]: [string, any]) => {
		const slug = path
			.split('/')
			.pop()
			?.replace(/\.(md|svx)$/, '');
		return {
			slug,
			metadata: {
				title: post.metadata.title,
				date: post.metadata.date,
				description: post.metadata.description,
				tags: post.metadata.tags || []
			}
		};
	});

	const tags = posts.reduce(
		(acc, post) => {
			(post.metadata.tags || []).forEach((tag: string) => {
				if (!acc[tag]) {
					acc[tag] = [];
				}
				acc[tag].push(post);
			});
			return acc;
		},
		{} as Record<string, TagPost[]>
	);

	return {
		tags,
		postsCount: posts.length
	};
};
