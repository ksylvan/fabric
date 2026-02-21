<script lang="ts">
  import { onMount } from 'svelte';
  import { noteStore } from '$lib/store/note-store';
  import { drawerStore } from '$lib/store/drawer-store';
  import { toastStore } from '$lib/store/toast-store';
  import { clickOutside } from '$lib/actions/clickOutside';

  let textareaEl: HTMLTextAreaElement;
  let saving = false;

  let content = '';

  // Auto-resize textarea
  function adjustTextareaHeight() {
    if (textareaEl) {
      textareaEl.style.height = 'auto';
      textareaEl.style.height = textareaEl.scrollHeight + 'px';
    }
  }

  async function saveContent() {
    if (!$noteStore.content.trim()) {
      toastStore.info('Cannot save empty note');
      return;
    }

    try {
      saving = true;
      await noteStore.save();

      toastStore.success('Note saved successfully!');
    } catch (error) {
      console.error('Failed to save note:', error);
      toastStore.error(error instanceof Error ? error.message : 'Failed to save notes');
    } finally {
      saving = false;
    }
  }

  function closeDrawer() {
    if ($noteStore.isDirty) {
      if (confirm('You have unsaved changes. Are you sure you want to close?')) {
        noteStore.reset();
        drawerStore.close();
      }
    } else {
      drawerStore.close();
    }
  }

  // Load saved content when drawer opens
  $: if ($drawerStore) {
    const savedContent = localStorage.getItem('savedText');
    if (savedContent) {
      noteStore.updateContent(savedContent);
      noteStore.save();
    }
  }

  // Keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && event.key === 's') {
      event.preventDefault();
      saveContent();
    }
    if (event.key === 'Escape') {
      closeDrawer();
    }
  }

  onMount(() => {
    adjustTextareaHeight();
  });
</script>

{#if $drawerStore}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/50 z-40 transition-opacity"
    on:click={closeDrawer}
    on:keydown={handleKeydown}
    role="button"
    tabindex="-1"
  ></div>

  <!-- Drawer panel -->
  <aside
    class="fixed top-0 right-0 h-full w-[40%] bg-surface-900 z-50 shadow-xl flex flex-col p-4 pt-20 transition-transform"
    use:clickOutside={closeDrawer}
  >
    <div class="flex flex-col h-full">
      <header class="flex-none p-2 border-b border-white/10">
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold">Notes</h2>
          <button
            class="text-white/70 hover:text-white transition-colors text-xl leading-none"
            on:click={closeDrawer}
            aria-label="Close drawer"
          >&times;</button>
        </div>
        <div class="flex justify-between items-center mt-1">
          {#if $noteStore.lastSaved}
            <span class="text-xs opacity-70">
              Last saved: {$noteStore.lastSaved.toLocaleTimeString()}
            </span>
          {/if}
        </div>
        <div class="flex gap-4 mt-2 text-xs opacity-70">
          <span>Notes saved to <code>inbox/</code></span>
          <span>Ctrl + S to save</span>
        </div>
      </header>

      <div class="flex-1 p-2">
        <textarea
          bind:this={textareaEl}
          value={$noteStore.content}
          on:input={e => noteStore.updateContent(e.currentTarget.value)}
          on:keydown={handleKeydown}
          class="w-full h-full min-h-[300px] resize-none p-2 rounded-lg bg-primary-800/30 border-none text-sm"
          placeholder="Enter your text here..."
        />
      </div>

      <footer class="flex-none flex justify-between items-center p-2 border-t border-white/10">
        <span class="text-xs opacity-70">
          {#if $noteStore.isDirty}
            Unsaved changes
          {/if}
        </span>
        <div class="flex gap-2">
          <button
            class="px-3 py-1.5 text-sm rounded-md bg-white/10 hover:bg-white/20 transition-colors"
            on:click={noteStore.reset}
          >
            Reset
          </button>
          <button
            class="px-3 py-1.5 text-sm rounded-md bg-primary-500 hover:bg-primary-600 text-white transition-colors"
            on:click={saveContent}
          >
            {#if saving}
              Saving...
            {:else}
              Save
            {/if}
          </button>
        </div>
      </footer>
    </div>
  </aside>
{/if}
