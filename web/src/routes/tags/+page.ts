import type { PageLoad } from './$types';
import type { PostSummary } from '$lib/components/posts/post-interface';

export const load: PageLoad = async () => {
	const postFiles = import.meta.glob('/src/lib/content/posts/*.{md,svx}', { eager: true });

	const posts: PostSummary[] = Object.entries(postFiles).map(([path, post]: [string, any]) => {
		const slug =
			path
				.split('/')
				.pop()
				?.replace(/\.(md|svx)$/, '') ?? '';
		return {
			slug,
			metadata: post.metadata
		};
	});

	const tags = posts.reduce(
		(acc, post) => {
			post.metadata.tags?.forEach((tag: string) => {
				if (!acc[tag]) {
					acc[tag] = [];
				}
				acc[tag].push(post);
			});
			return acc;
		},
		{} as Record<string, PostSummary[]>
	);

	return {
		tags,
		postsCount: posts.length
	};
};
