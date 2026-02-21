<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    sideOffset = 4,
    position = "popper",
    class: className = undefined,
    children,
    ...restProps
  }: {
    sideOffset?: number;
    position?: string;
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();
</script>

<div
  class={cn(
    "relative z-50 min-w-[8rem] overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-lg",
    position === "popper" &&
      "data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
    className
  )}
  {...restProps}
>
  <div
    class={cn(
      "p-1",
      position === "popper" &&
        "w-full min-w-[var(--radix-select-trigger-width)]"
    )}
  >
    {#if children}{@render children()}{/if}
  </div>
</div>
