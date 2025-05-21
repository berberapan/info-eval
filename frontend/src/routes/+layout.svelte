<script lang="ts">
  import { goto } from '$app/navigation';
  import '../app.css';
</script>

<div class="app-container min-h-screen flex flex-col">
  <header class="bg-neutral text-neutral-content p-4 shadow-md">
    <nav class="container mx-auto flex justify-between items-center">
    </nav>
  </header>
  <main class="flex-grow container mx-auto p-4">
    <slot />
  </main>
</div>

