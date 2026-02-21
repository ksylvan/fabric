<script lang="ts">
  //import Search from './Search.svelte';
  import type { PageData } from './$types';
  import Card from '$lib/components/ui/cards/card.svelte';
  import { Youtube } from 'svelte-youtube-lite';
  import PostCard from '$lib/components/posts/PostCard.svelte';
  import Connections from '$lib/components/ui/connections/Connections.svelte';
  import Button from '$lib/components/ui/button/button.svelte';

  let searchQuery = '';
  let selectedTags: string[] = [];
  let allTags: string[] = [];
  let inputValue = '';

  export let data: PageData;
  $: posts = data.posts || [];

  // Extract all unique tags from posts
  $: {
    const tagSet = new Set<string>();
    posts?.forEach(post => {
      post.metadata?.tags?.forEach(tag => tagSet.add(tag));
    });
    allTags = Array.from(tagSet);
  }

  // Filter posts based on selected tags
  $: filteredPosts = posts?.filter(post => {
    if (selectedTags.length === 0) return true;
    return selectedTags.every(tag => 
      post.metadata?.tags?.some(postTag => postTag.toLowerCase() === tag.toLowerCase())
    );
  }) || [];

  // Filter posts based on search query
  $: searchResults = filteredPosts.filter(post => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      post.metadata?.title?.toLowerCase().includes(query) ||
      post.metadata?.description?.toLowerCase().includes(query) ||
      post.metadata?.tags?.some(tag => tag.toLowerCase().includes(query))
    );
  });

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

  function handleChipKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    }
    if (event.key === 'Backspace' && !inputValue && selectedTags.length > 0) {
      selectedTags = selectedTags.slice(0, -1);
    }
  }
</script>

<!-- <Search /> -->

<div class="absolute inset-0 -z-10 overflow-hidden h-96">
  <Connections  particleCount={100} particleSize={3} particleSpeed={0.1} connectionDistance={100}/>
</div>

<div class="py-12">
  <h1 class="mb-4 text-3xl font-bold">Knowledge Garden</h1>
  <p class="text-sm mb-4 font-small">A digital space where ideas grow and connections flourish</p>

  <div class="mx-auto ml-auto grid grid-cols-1 md:grid-cols-2 gap-4 justify-end">
    <div class="container mx-auto justify-left">
      <img src="https://img.shields.io/github/languages/top/danielmiessler/fabric" alt="Github top language">
      <img src="https://img.shields.io/github/last-commit/danielmiessler/fabric" alt="GitHub last commit">
      <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
      <br>
      <hr class="!border-t-4" />
      <br>
      <h4 class="text-lg font-bold"><b>Leverage Proven Patterns</b></h4>
      <br>
      <Youtube id="UbDyjIIGaxQ" title="Network Chuck Explains fabric" />
      <br>
        <p>Leverage the power of the patterns. Use AI assistance to amplify your creativity. The templates are designed to 
          help you focus on what matters most - your ideas. Start with structured frameworks, then make them your own.
        </p>
      <br>

    </div>
    <div class="container mx-auto ml-auto grid grid-cols-1 md:grid-cols-2 gap-4 m-4">

      <div>
        <h4 class=""><b>Find your interests, build your knowledge</b></h4>
      </div>
      <div class="container mx-auto md:col-start-1 pt-4">
        <p>Embark on an enriching journey of self-discovery through the power of words! Sharing your unique voice and experiences isn't just 
          about expressing yourself; it's about connecting, inspiring, and empowering others with your story.
        </p>
        <br>
        <p>Regular writing is more than just a means to share; it's a tool that deepens your self-awareness, helping you understand yourself 
          better and grow in the process. 
        </p>
      </div>
      <div class="md:col-start-2">
        <br>
        <Card
          header="Let Your Voice Be Heard"
          imageUrl="/brain.png"
          imageAlt="Blog post header image"
          title="Welcome to Your Digital Garden"
          content="Start creating, connecting, and sharing your knowledge"
          authorName="Your Name Here"
          authorAvatarUrl="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiIGNsYXNzPSJsdWNpZGUgbHVjaWRlLXVzZXIiPjxwYXRoIGQ9Ik0xOSAyMXYtMmE0IDQgMCAwIDAtNC00SDlhNCA0IDAgMCAwLTQgNHYyIi8+PGNpcmNsZSBjeD0iMTIiIGN5PSI3IiByPSI0Ii8+PC9zdmc+"
          link="/posts/welcome"
        />
      </div>
    </div>
    <div class="container mx-auto ml-auto grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
      <Card
        header="Shape Your Ideas"
        imageUrl="/electric.png"
        imageAlt="Blog post header image"
        title="Transform Knowledge into Action and Insight"
        content="Create, Connect, and Share Your Knowledge"
        authorName="Your Name Here"
        authorAvatarUrl="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiIGNsYXNzPSJsdWNpZGUgbHVjaWRlLXVzZXIiPjxwYXRoIGQ9Ik0xOSAyMXYtMmE0IDQgMCAwIDAtNC00SDlhNCA0IDAgMCAwLTQgNHYyIi8+PGNpcmNsZSBjeD0iMTIiIGN5PSI3IiByPSI0Ii8+PC9zdmc+"
        link="/tags/template"
      />
      <div class="container mx-auto justify-right">
        <blockquote class="border-l-4 border-primary-500 pl-4 italic text-muted-foreground">
          There are many patterns for different use cases. How will you use them to your advantage?
        </blockquote>
        <br>
        <p>AI isn't just a tool - it's your creative companion. Use it to explore ideas, generate outlines, or refine your writing. 
          But remember, the authentic voice, the unique insights, and the valuable experiences - those come from you. This is where 
          technology meets creativity to help you build something truly meaningful.
        </p>
      </div>
    </div>
    <div class="container mx-auto ml-auto grid grid-cols-1 md:grid-cols-2 gap-4 justify-end max-h-36 mt-8 pb-8">
      <div class="md:col-start-1">
        <!-- This card should be replaced with explainer graphic or text -->
        <Card
          header="Backed by Obsidian"
          imageUrl="/obsidian-logo.png"
          imageAlt="Blog post header image"
          title="Connected Thinking"
          content="Build your knowledge network"
          authorName="Your Name Here"
          authorAvatarUrl="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiIGNsYXNzPSJsdWNpZGUgbHVjaWRlLXVzZXIiPjxwYXRoIGQ9Ik0xOSAyMXYtMmE0IDQgMCAwIDAtNC00SDlhNCA0IDAgMCAwLTQgNHYyIi8+PGNpcmNsZSBjeD0iMTIiIGN5PSI3IiByPSI0Ii8+PC9zdmc+"
          link="/posts/obsidian"
        /> 
      </div>
      <div class="container mx-auto md:col-start-2 justify-left">
        <hr class="!border-t-4" />
        <br>
        <h4 class="text-lg font-bold">Build Your Knowledge Network • Share Your Journey • Inspire Others</h4>
      </div>
    </div>
  </div>
  <br>
  <div class="rounded-tl-lg m-auto grid grid-cols-1 gap-4 mt-8">
    <div class="mx-auto max-h-52 max-w-52"><img src="/fabric-logo.png" alt="fabric-logo"></div>
  </div>
  <br>
  <div class="container mx-auto justify-center grid grid-cols-1 gap-4 mt-8">
      <hr class="!border-t-4" />
      <br>
      <h4 class="text-lg font-bold">Showcase your interests. Tell people what you've been working on. Create your community.</h4>
  </div>
</div>
<div class="container py-12">
  <div class="my-4">
    <div class="flex flex-wrap items-center gap-2 rounded-md border border-white/20 bg-transparent px-3 py-2 focus-within:ring-2 focus-within:ring-primary-500">
      {#each selectedTags as tag}
        <span class="inline-flex items-center gap-1 rounded-full bg-primary-500/20 px-2.5 py-0.5 text-sm text-primary-300">
          {tag}
          <button
            type="button"
            class="ml-1 text-primary-300 hover:text-white"
            on:click={() => removeTag(tag)}
          >&times;</button>
        </span>
      {/each}
      <input
        type="text"
        name="tags"
        placeholder="Filter by tags..."
        class="flex-1 min-w-[120px] bg-transparent border-none outline-none text-sm placeholder:text-white/50"
        bind:value={inputValue}
        on:keydown={handleChipKeydown}
      />
    </div>
  </div>
  <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
    {#each searchResults as post}
      <PostCard {post} /> <!-- TODO: Add images to post metadata --> 
    {/each}
  </div>
</div>
