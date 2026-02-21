<script lang="ts">
	import type { PageData } from './$types';
	import PostContent from '$lib/components/posts/PostContent.svelte';

	let { data }: { data: PageData } = $props();
</script>

<PostContent
	post={{
		content: data.content,
		metadata: data.metadata,
		slug: data.slug
	}}
/>
