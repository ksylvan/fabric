<script lang="ts">
	import type { PageData } from './$types';
	import PostContent from '$lib/components/posts/PostContent.svelte';

	export let data: PageData;
</script>

<PostContent
	post={{
		content: data.content,
		metadata: data.metadata,
		slug: data.slug
	}}
/>
