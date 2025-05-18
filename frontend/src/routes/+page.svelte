<script>
  import { goto } from '$app/navigation'; 

  function goToScenarios() {
    goto('/scenarios');   
    }
</script>

<div class="flex flex-col items-center justify-center min-h-screen bg-base-200 p-4">
  <div class="card w-full max-w-md bg-base-100 shadow-xl">
    <div class="card-body items-center text-center">
      <h1 class="card-title text-3xl font-bold mb-4">Välkommen!</h1>
      <p class="mb-6">
        Klicka på knappen nedan för att se dina uppgifter.
      </p>
      <div class="card-actions">
        <button
          on:click={goToScenarios}
          class="btn btn-primary btn-lg"
        >
          Uppgifter
        </button>
      </div>
    </div>
  </div>
</div>
