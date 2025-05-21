<script lang="ts">
  import { goto } from '$app/navigation';

  let email = "";
  let password = "";
  let errorMessage = "";
  let isLoading = false;

  async function handleLogin(): Promise<void> {
    isLoading = true;
    errorMessage = "";

    try {
      const response = await fetch('http://localhost:9000/v1/authentication', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include', 
        body: JSON.stringify({ email, password })
      });

      if (response.ok) {
        goto('/scenarios'); 
      } else {
        const errorData = await response.json().catch(() => ({ message: `Login failed (status: ${response.status})` }));
        errorMessage = errorData.error?.message || errorData.message || `Login failed (status: ${response.status})`;
      }
    } catch (error: any) {
      errorMessage = error.message || "An error occurred. Please try again later.";
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200 p-4">
  <div class="card w-full max-w-sm shadow-2xl bg-base-100">
    <form on:submit|preventDefault={handleLogin} class="card-body">
      <h2 class="text-2xl font-bold text-center mb-6">Login</h2>

      {#if errorMessage}
        <div class="alert alert-error shadow-lg mb-4">
          <div>
            <span>{errorMessage}</span>
          </div>
        </div>
      {/if}

      <div class="form-control">
        <label class="label" for="email-input">
          <span class="label-text">Email</span>
        </label>
        <input
          type="email"
          id="email-input"
          bind:value={email}
          class="input input-bordered"
          required
          disabled={isLoading}
        />
      </div>

      <div class="form-control mt-4">
        <label class="label" for="password-input">
          <span class="label-text">Password</span>
        </label>
        <input
          type="password"
          id="password-input"
          bind:value={password}
          class="input input-bordered"
          required
          disabled={isLoading}
        />
      </div>

      <div class="form-control mt-8">
        <button type="submit" class="btn btn-primary" disabled={isLoading}>
          {#if isLoading}
            <span class="loading loading-spinner"></span>
            Logging in...
          {:else}
            Login
          {/if}
        </button>
      </div>
    </form>
  </div>
</div>

