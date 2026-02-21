<script lang="ts">
  import type { PostMetadata } from '$lib/types';
  import type { Post } from '$lib/interfaces/post-interface'
  import PostCard from '$lib/components/posts/PostCard.svelte';

  let searchQuery = $state('');
  let selectedTags = $state<string[]>([]);
  let inputValue = $state('');

  let data: PageData;
  let posts = data.posts || [];

  // Extract all unique tags from posts
  let allTags = $derived.by(() => {
    const tagSet = new Set<string>();
    posts?.forEach(post => {
      post.metadata?.tags?.forEach(tag => tagSet.add(tag));
    });
    return Array.from(tagSet);
  });

  // Filter posts based on selected tags
  let filteredPosts = $derived(
    posts?.filter(post => {
      if (selectedTags.length === 0) return true;
      return selectedTags.every(tag =>
        post.metadata?.tags?.some(postTag => postTag.toLowerCase() === tag.toLowerCase())
      );
    }) || []
  );

  // Filter posts based on search query
  let searchResults = $derived(
    filteredPosts.filter(post => {
      if (!searchQuery) return true;
      const query = searchQuery.toLowerCase();
      return (
        post.metadata?.title?.toLowerCase().includes(query) ||
        post.metadata?.description?.toLowerCase().includes(query) ||
        post.metadata?.tags?.some(tag => tag.toLowerCase().includes(query))
      );
    })
  );

  function validateTag(value: string): boolean {
    return allTags.some(tag => tag.toLowerCase() === value.toLowerCase());
  }

  function addTag() {
    const value = inputValue.trim();
    if (value && validateTag(value) && !selectedTags.includes(value)) {
      selectedTags = [...selectedTags, value];
      inputValue = '';
    }
  }

  function removeTag(tag: string) {
    selectedTags = selectedTags.filter(t => t !== tag);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    }
    if (event.key === 'Backspace' && !inputValue && selectedTags.length > 0) {
      selectedTags = selectedTags.slice(0, -1);
    }
  }
</script>


<div class="container py-12">
	<div class="my-4">
		<div class="flex flex-wrap items-center gap-2 rounded-md border border-white/20 bg-transparent px-3 py-2 focus-within:ring-2 focus-within:ring-primary-500">
			{#each selectedTags as tag}
				<span class="inline-flex items-center gap-1 rounded-full bg-primary-500/20 px-2.5 py-0.5 text-sm text-primary-300">
					{tag}
					<button
						type="button"
						class="ml-1 text-primary-300 hover:text-white"
						onclick={() => removeTag(tag)}
					>&times;</button>
				</span>
			{/each}
			<input
				type="text"
				name="tags"
				placeholder="Filter by tags..."
				class="flex-1 min-w-[120px] bg-transparent border-none outline-none text-sm placeholder:text-white/50"
				bind:value={inputValue}
				onkeydown={handleKeydown}
			/>
		</div>
  </div>
</div>
<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
	{#each searchResults as post}
		<PostCard {post} /> <!-- TODO: Add images to post metadata -->
	{/each}
 </div>

