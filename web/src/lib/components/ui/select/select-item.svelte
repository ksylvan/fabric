<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    value,
    disabled = false,
    class: className = undefined,
    children,
    ...restProps
  }: {
    value: any;
    disabled?: boolean;
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();
</script>

<option
  {value}
  {disabled}
  class={cn("relative flex w-full cursor-default select-none py-1.5 pl-2 pr-2 text-sm outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50", className)}
  {...restProps}
>
  {#if children}{@render children()}{/if}
</option>
