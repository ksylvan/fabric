<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    placeholder = undefined,
    class: className = undefined,
    children,
    ...restProps
  }: {
    placeholder?: string;
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();
</script>

<span
  class={cn("text-sm", className)}
  {...restProps}
>
  {#if children}
    {@render children()}
  {:else if placeholder}
    {placeholder}
  {/if}
</span>
