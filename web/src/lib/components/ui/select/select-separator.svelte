<script lang="ts">
  import { cn } from "$lib/utils/utils";

  let {
    class: className = undefined,
    ...restProps
  }: {
    class?: string;
    [key: string]: any;
  } = $props();
</script>

<div class={cn("bg-muted -mx-1 my-1 h-px", className)} {...restProps} />
