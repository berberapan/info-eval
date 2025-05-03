<script>
  let email = "";
  let password = "";

  function handleLogin() {
    console.log("Logging in with", email, password);
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="card w-full max-w-sm shadow-2xl bg-base-100">
    <form on:submit|preventDefault={handleLogin} class="card-body">
      <h2 class="text-2xl font-bold text-center mb-4">Login</h2>

      <div class="form-control">
        <label class="label" for="email-input">
          <span class="label-text">Email</span>
        </label>
        <input
          type="email"
          placeholder="email"
          bind:value={email}
          class="input input-bordered"
          required
        />
      </div>

      <div class="form-control">
        <label class="label" for="password-input">
          <span class="label-text">Password</span>
        </label>
        <input
          type="password"
          placeholder="password"
          bind:value={password}
          class="input input-bordered"
          required
        />
      </div>

      <div class="form-control mt-6">
        <button type="submit" class="btn btn-primary">
          Login
        </button>
      </div>
    </form>
  </div>
</div>

