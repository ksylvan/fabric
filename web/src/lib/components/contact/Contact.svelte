<script lang="ts">
  import { Contact } from 'lucide-svelte';
</script>

<div class="form-control w-full m-auto p-4 rounded-lg bg-gradient-to-br from-green-500 to-amber-400 shadow-lg text-current" title="contact form">
  <h2 class="font-bold pl-2">We'd love to hear from you</h2>
  <p class="font-bold pl-2">Email</p>
  <div class="flex items-center border border-white/20 rounded-md overflow-hidden divide-x divide-white/20">
    <input type="text" class="flex-1 bg-transparent px-3 py-2 text-sm outline-none" placeholder="Enter an email address where you can be reached..." />
  </div>
  <p class="font-bold pl-2">Website</p>
  <div class="flex items-center border border-white/20 rounded-md overflow-hidden divide-x divide-white/20">
    <div class="px-3 py-2 bg-white/10 text-sm">https://</div>
    <input type="text" class="flex-1 bg-transparent px-3 py-2 text-sm outline-none" placeholder="www.example.com" />
  </div>
  <p class="font-bold pl-2">Contact Information</p>
  <div class="flex items-center border border-white/20 rounded-md overflow-hidden divide-x divide-white/20">
    <input type="text" class="flex-1 bg-transparent px-3 py-2 text-sm outline-none" placeholder="Enter a other contact information here..." />
  </div>
  <label class="label">
    <span class="font-bold pl-2">Message</span>
    <textarea class="w-full rounded-md border border-white/20 bg-transparent px-3 py-2 text-sm outline-none" rows="4" placeholder="Enter your message ..." />
  </label>
  <a href="/" title=""><button class="rounded-lg bg-secondary-500 px-4 py-2 text-sm font-medium text-white hover:bg-secondary-600 transition-colors"><Contact /></button></a>

</div>
