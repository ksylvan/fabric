<script lang="ts">
  import { cn } from "$lib/utils/utils";
  import { buttonVariants } from "./index.js";

  let {
    variant = "",
    size = "default",
    class: className = undefined,
    children,
    ...restProps
  }: {
    variant?: string;
    size?: string;
    class?: string;
    children?: import('svelte').Snippet;
    [key: string]: any;
  } = $props();

  let classes = $derived(cn(
    buttonVariants.base,
    buttonVariants.variants.variant[variant as keyof typeof buttonVariants.variants.variant],
    buttonVariants.variants.size[size as keyof typeof buttonVariants.variants.size],
    className
  ));
</script>

<button
	class={classes}
	type="button"
	{...restProps}
>
	{#if children}{@render children()}{/if}
</button>
