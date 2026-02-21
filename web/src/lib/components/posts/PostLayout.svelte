<script>
  import { type Snippet } from 'svelte';

  let { aliases, date, tags, title, description, author, updated, children } = $props();
</script>

<article class="prose prose-slate dark:prose-invert max-w-5xl flex-1">
  {#if aliases}
  <h1 class="inline-block text-4xl font-bold inherit-colors lg:text-5xl">{aliases}</h1>
  {/if}
  {@render children?.()}
</article>

<style lang="postcss">

:global(h1) {
  @apply h1; 
}

:global(h2) {
  @apply h2; 
}

:global(p) {
  @layer p; 
}

:global(ul) {
  @layer ul; 
}

:global(li) {
  @layer li; 
}

:global(a) {
  @layer a; 
}

:global(img) {
  @layer img; 
}

:global(blockquote) {
  @layer blockquote; 
}

:global(code) {
  @layer code; 
}

</style> 
