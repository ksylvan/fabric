<script>
  let { header = '', imageUrl = '', imageAlt = 'Post', title = '', content = '', authorName = '', authorAvatarUrl = '', link = '' } = $props();
</script>

<div class="w-full grid grid-cols-1 md:grid-cols-1 justify-end gap-4">
  <a class="block overflow-hidden rounded-lg border bg-gray-800/50 shadow-sm transition-colors hover:bg-muted/50" href={link}>
    <header>
      <img src={imageUrl} class="bg-black/50 w-full" alt={imageAlt} />
    </header>
    <div class="p-4 space-y-4">
      <h6 class="text-sm font-semibold" data-toc-ignore>{header}</h6>
      <h3 class="text-xl font-bold" data-toc-ignore>{title}</h3>
      <article>
        <p>
          {content}
        </p>
      </article>
    </div>
    <hr class="opacity-50" />
    <footer class="p-4 flex justify-start items-center space-x-4">
      <img src={authorAvatarUrl} alt="{authorName} avatar" class="w-8 h-8 rounded-full object-cover" />
      <div class="flex-auto flex justify-between items-center">
        <h6 class="font-bold" data-toc-ignore>By {authorName}</h6>
        <small>On {new Date().toLocaleDateString()}</small>
      </div>
    </footer>
  </a>
</div>
