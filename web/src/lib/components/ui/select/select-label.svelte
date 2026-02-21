<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    class: className = undefined,
    children,
    ...restProps
  }: {
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();
</script>

<span
  class={cn("px-2 py-1.5 text-sm bg-secondary-500 text-white rounded font-semibold", className)}
  {...restProps}
>
  {#if children}{@render children()}{/if}
</span>
