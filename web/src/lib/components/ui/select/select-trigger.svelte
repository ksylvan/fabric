<script lang="ts">
  import { ChevronDown } from "lucide-svelte";
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

<button
  type="button"
  class={cn(
    "flex h-9 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
    className
  )}
  {...restProps}
>
  {#if children}{@render children()}{/if}
  <ChevronDown class="h-4 w-4 opacity-50" />
</button>
