<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation'; // Import goto for navigation

  // Reactive variable to store the scenarios
  let scenarios = [];
  let isLoading = true; // To show a loading state
  let error = null; // To display any errors during fetch

  const API_URL = 'http://localhost:9000/v1/scenarios'; 
  // Function to fetch scenarios from the API
  async function fetchScenarios() {
    isLoading = true;
    error = null;
    try {
      const response = await fetch(API_URL); // Fetch data from the API_URL

      if (!response.ok) {
        // If the response is not OK (e.g., 404, 500), throw an error
        const errorData = await response.json().catch(() => ({ message: response.statusText }));
        throw new Error(`HTTP error! Status: ${response.status} - ${errorData.message || 'Failed to fetch'}`);
      }

      const data = await response.json(); // Parse the JSON response

      // Assuming the API returns an object with a 'scenarios' array
      if (data && data.scenarios) {
        scenarios = data.scenarios;
      } else {
        // Handle cases where the data format might be different or 'scenarios' is missing
        console.warn("API response did not contain a 'scenarios' array, or data is malformed. Received:", data);
        scenarios = Array.isArray(data) ? data : []; // Fallback if the root is an array
      }

    } catch (e) {
      console.error("Failed to fetch scenarios:", e);
      error = `Kunde inte ladda scenarier: ${e.message}`; // "Could not load scenarios: [error message]"
      scenarios = []; // Clear scenarios on error
    } finally {
      isLoading = false;
    }
  }

  // Fetch scenarios when the component is mounted
  onMount(() => {
    fetchScenarios();
  });

  // Placeholder function for editing a scenario
  function editScenario(id) {
    console.log("Edit scenario with ID:", id);
    // Navigate to an edit page, e.g., using SvelteKit's goto
    // goto(`/uppgifter/edit/${id}`);
    alert(`Redigera scenario: ${id}`); // Placeholder
  }

  // Placeholder function for sending a scenario
  function sendScenario(id) {
    console.log("Send scenario with ID:", id);
    // Implement sending logic, e.g., another API call
    alert(`Skicka scenario: ${id}`); // Placeholder
  }

  // Placeholder function for creating a new scenario
  function createNewScenario() {
    console.log("Create new scenario");
    // Navigate to a creation page
    // goto('/uppgifter/new');
    alert("Skapa nytt scenario"); // Placeholder
  }

  // Helper to get difficulty text
  function getDifficultyText(level) {
    switch (level) {
      case 1: return "Lätt"; // Easy
      case 2: return "Medel"; // Medium
      case 3: return "Svår"; // Hard
      default: return "Okänd"; // Unknown
    }
  }
</script>

<div class="container mx-auto p-4 md:p-8 min-h-screen bg-base-200">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-3xl md:text-4xl font-bold text-base-content">Scenarier</h1>
    <button on:click={createNewScenario} class="btn btn-accent">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
      </svg>
      Skapa Nytt Scenario
    </button>
  </div>

  {#if isLoading}
    <div class="flex justify-center items-center h-64">
      <span class="loading loading-lg loading-spinner text-primary"></span>
      <p class="ml-4 text-xl">Laddar scenarier...</p>
    </div>
  {:else if error}
    <div role="alert" class="alert alert-error">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>{error}</span>
      <button on:click={fetchScenarios} class="btn btn-sm btn-ghost ml-auto">Försök igen</button>
    </div>
  {:else if scenarios.length === 0}
    <div class="text-center py-10">
      <p class="text-xl text-base-content mb-4">Inga scenarier hittades.</p>
      <button on:click={fetchScenarios} class="btn btn-primary">Ladda om</button>
    </div>
  {:else}
    <div class="card w-full bg-base-100 shadow-xl overflow-x-auto">
      <div class="card-body p-0 md:p-4">
        <table class="table table-zebra w-full">
          <thead>
            <tr>
              <th class="w-1/4 md:w-1/5">Titel</th>
              <th class="w-1/12 md:w-1/12 text-center">Svårighetsgrad</th>
              <th class="w-2/5 md:w-2/5">Beskrivning</th>
              <th class="w-1/4 md:w-1/5 text-center">Åtgärder</th>
            </tr>
          </thead>
          <tbody>
            {#each scenarios as scenario (scenario.id)}
              <tr>
                <td class="font-semibold align-top">{scenario.title}</td>
                <td class="text-center align-top">
                  <span class:badge-success={scenario.difficulty === 1}
                        class:badge-warning={scenario.difficulty === 2}
                        class:badge-error={scenario.difficulty === 3}
                        class="badge badge-ghost badge-md">
                    {getDifficultyText(scenario.difficulty)}
                  </span>
                </td>
                <td class="text-sm text-gray-600 align-top whitespace-normal break-words">
                  {scenario.description}
                </td>
                <td class="text-center align-top">
                  <div class="flex flex-col sm:flex-row justify-center items-center gap-2">
                    <button
                      on:click={() => editScenario(scenario.id)}
                      class="btn btn-sm btn-outline btn-info w-full sm:w-auto"
                    >
                      Redigera
                    </button>
                    <button
                      on:click={() => sendScenario(scenario.id)}
                      class="btn btn-sm btn-outline btn-success w-full sm:w-auto"
                    >
                      Skicka
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Ensure DaisyUI is imported in your global CSS file (e.g., app.css) */
  .container {
    max-width: 1200px; /* Or your preferred max width for the content area */
  }
  .table td, .table th {
    padding: 0.75rem; /* Adjust padding as needed */
  }
  .table td.whitespace-normal {
    white-space: normal; /* Ensure long descriptions wrap */
  }
</style>
