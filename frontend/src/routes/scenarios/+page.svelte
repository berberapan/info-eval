<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation'; 
  import { authStore, checkAuthClient } from '$lib/stores/auth'; 

  let scenarios = [];
  let isLoading = true; 
  let error = null; 
  let magicLinkInfo = null; 
  let isSending = {}; 

  const SCENARIOS_API_URL = 'http://localhost:9000/v1/scenarios';
  const SESSIONS_API_URL = 'http://localhost:9000/v1/sessions'; 

  async function fetchScenarios() {
    isLoading = true;
    error = null;
    magicLinkInfo = null; 
    try {
      const response = await fetch(SCENARIOS_API_URL);
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: response.statusText }));
        throw new Error(`HTTP error! Status: ${response.status} - ${errorData.message || 'Failed to fetch scenarios'}`);
      }
      const data = await response.json();
      if (data && data.scenarios) {
        scenarios = data.scenarios;
      } else {
        console.warn("API response did not contain a 'scenarios' array, or data is malformed. Received:", data);
        scenarios = Array.isArray(data) ? data : [];
      }
    } catch (e) {
      console.error("Failed to fetch scenarios:", e);
      error = `Kunde inte ladda övningar: ${e.message}`;
      scenarios = [];
    } finally {
      isLoading = false;
    }
  }
  onMount(async () => {
    await checkAuthClient();
    await fetchScenarios();
  });

  async function sendScenario(scenarioId) {
    isSending = { ...isSending, [scenarioId]: true };
    magicLinkInfo = null; // Clear previous link first
    if ($authStore.isLoading) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    if (!$authStore.isAuthenticated) {
      error = "Du måste vara inloggad för att skapa en delningslänk. Vänligen logga in.";
      isSending = { ...isSending, [scenarioId]: false };
      await checkAuthClient();
      return;
    }
    error = null; 
    try {
      const payload = {
        scenario_id: scenarioId,
        notes: `Session skapad för scenario ${scenarioId} den ${new Date().toLocaleString()}`,
        validity_duration_hours: 24 
      };
      const response = await fetch(SESSIONS_API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', 
        body: JSON.stringify(payload)
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: response.statusText }));
        let errorMessage = `HTTP error! Status: ${response.status}`;
        if (errorData.error && typeof errorData.error === 'string') {
            errorMessage += ` - ${errorData.error}`;
        } else if (errorData.error && errorData.error.message) {
            errorMessage += ` - ${errorData.error.message}`;
        } else if (errorData.message) {
            errorMessage += ` - ${errorData.message}`;
        } else {
            errorMessage += ` - ${response.statusText || 'Failed to create session'}`;
        }
        if (response.status === 401 || response.status === 403) {
            errorMessage += " Din session kan ha gått ut, vänligen logga in igen.";
            await checkAuthClient(); 
        }
        throw new Error(errorMessage);
      }
      const result = await response.json();
      if (result && result.scenario_session) {
        const session = result.scenario_session;
        const studentLink = `${window.location.origin}/session/${session.id}`;
        const teacherResultsLink = `${window.location.origin}/teacher/session-results/${session.id}`;
        magicLinkInfo = {
          id: session.id, 
          token: session.token, 
          expiresAt: new Date(session.expires_at).toLocaleString(),
          studentLink: studentLink,
          teacherResultsLink: teacherResultsLink, 
          scenarioTitle: scenarios.find(s => s.id === scenarioId)?.title || scenarioId,
          message: `Magic link skapad för övning: '${scenarios.find(s => s.id === scenarioId)?.title || scenarioId}'!`
        };
        error = null; 
      } else {
        throw new Error("Session data not found or malformed in API response.");
      }
    } catch (e) {
      console.error("Failed to create magic link:", e);
      error = `Kunde inte skapa magic link: ${e.message}`;
      magicLinkInfo = null;
    } finally {
      isSending = { ...isSending, [scenarioId]: false };
    }
  }
  
  function getDifficultyText(level) {
    switch (level) {
      case 1: return "1";
      case 2: return "2";
      case 3: return "3";
      default: return "Extra uppgift";
    }
  }

  function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
      alert('Länk kopierad!');
    }).catch(err => {
      console.error('Kunde inte kopiera text: ', err);
      alert('Kunde inte kopiera länk. Försök manuellt.');
    });
  }
</script>

<div class="container mx-auto p-4 md:p-8 min-h-screen bg-base-200 text-base-content">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-3xl md:text-4xl font-bold">Övningar</h1>
    {#if $authStore.isAuthenticated}
      <button on:click={createNewScenario} class="btn btn-accent">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
        Skapa ny övning
      </button>
    {/if}
  </div>

  {#if $authStore.isLoading}
    <div class="alert alert-info my-4 shadow-md">
      <span>Kontrollerar inloggningsstatus...</span>
    </div>
  {:else if !$authStore.isAuthenticated}
    <div class="alert alert-warning my-4 shadow-md">
      <span>Du är inte inloggad. Vänligen <a href="/login" class="link link-primary">logga in</a> för att hantera scenarier och skapa delningslänkar.</span>
    </div>
  {/if}

  {#if isLoading}
    <div class="flex justify-center items-center h-64">
      <span class="loading loading-lg loading-spinner text-primary"></span>
      <p class="ml-4 text-xl">Laddar scenarier...</p>
    </div>
  {/if}

  {#if error && !isLoading}
    <div role="alert" class="alert alert-error my-4 shadow-lg">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <div>
        <h3 class="font-bold">Ett fel inträffade!</h3>
        <div class="text-xs">{error}</div>
      </div>
      {#if error.includes("Kunde inte ladda scenarier")}
        <button on:click={fetchScenarios} class="btn btn-sm btn-ghost ml-auto">Försök igen</button>
      {/if}
       <button class="btn btn-sm btn-ghost" on:click={() => error = null}>Stäng</button>
    </div>
  {/if}

  {#if magicLinkInfo && !error}
    <div role="alert" class="alert alert-success my-4 shadow-lg">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <div>
            <h3 class="font-bold">{magicLinkInfo.message}</h3>
            <div class="text-xs mt-1"><b>Session ID:</b> {magicLinkInfo.id}</div>
            <div class="text-xs mt-1">
                <b>Länk för studenter:</b>
                <div class="flex items-center gap-2">
                    <a href={magicLinkInfo.studentLink} target="_blank" rel="noopener noreferrer" class="link link-hover link-secondary font-semibold break-all">{magicLinkInfo.studentLink}</a>
                    <button class="btn btn-xs btn-ghost" title="Kopiera studentlänk" on:click={() => copyToClipboard(magicLinkInfo.studentLink)}>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v3.004c0 .414-.168.78-.468 1.062a2.25 2.25 0 01-1.062.468h-7.5a2.25 2.25 0 01-1.062-.468A2.25 2.25 0 013 8.25V5.25c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
                    </button>
                </div>
            </div>
            <div class="text-xs mt-1"><b>Giltig till:</b> {magicLinkInfo.expiresAt}</div>
            
            <div class="text-xs mt-2">
                <b>Resultat för denna session (lärare):</b>
                <div class="flex items-center gap-2">
                    <a href={magicLinkInfo.teacherResultsLink} class="btn btn-xs btn-outline btn-primary mt-1">
                        Visa Resultat
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 ml-1">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
                        </svg>
                    </a>
                </div>
            </div>
        </div>
        <button class="btn btn-sm btn-ghost" on:click={() => magicLinkInfo = null}>Stäng</button>
    </div>
  {/if}

  {#if !isLoading && !error && scenarios.length === 0}
    <div class="text-center py-10">
      <p class="text-xl mb-4">Inga scenarier hittades.</p>
      <button on:click={fetchScenarios} class="btn btn-primary">Ladda om scenarier</button>
    </div>
  {:else if !isLoading && scenarios.length > 0}
    <div class="card w-full bg-base-100 shadow-xl overflow-x-auto">
      <div class="card-body p-0 md:p-4">
        <table class="table table-zebra w-full">
          <thead>
            <tr>
              <th class="w-1/4 md:w-1/5">Titel</th>
              <th class="w-1/12 md:w-1/12 text-center">Årskurs</th>
              <th class="w-2/5 md:w-2/5">Beskrivning</th>
              <th class="w-1/4 md:w-1/5 text-center"></th>
            </tr>
          </thead>
          <tbody>
            {#each scenarios as scenario (scenario.id)}
              <tr>
                <td class="font-semibold align-top py-3 px-2 md:px-4">{scenario.title}</td>
                <td class="text-center align-top py-3 px-2 md:px-4">
                  <span class:badge-success={scenario.difficulty === 1}
                        class:badge-warning={scenario.difficulty === 2}
                        class:badge-error={scenario.difficulty === 3}
                        class="badge badge-ghost badge-md">
                    {getDifficultyText(scenario.difficulty)}
                  </span>
                </td>
                <td class="text-sm text-base-content/80 align-top py-3 px-2 md:px-4 whitespace-normal break-words">
                  {scenario.description || "Ingen beskrivning."}
                </td>
                <td class="text-center align-top py-3 px-2 md:px-4">
                  <div class="flex flex-col sm:flex-row justify-center items-center gap-2">
                    <button
                      on:click={() => editScenario(scenario.id)}
                      class="btn btn-sm btn-outline btn-info w-full sm:w-auto"
                      title="Redigera scenario {scenario.title}"
                      disabled={!$authStore.isAuthenticated}
                    >
                      Redigera
                    </button>
                    <button
                      on:click={() => sendScenario(scenario.id)}
                      class="btn btn-sm btn-outline btn-success w-full sm:w-auto"
                      disabled={isSending[scenario.id] || !$authStore.isAuthenticated}
                      title="Skapa och skicka en delningslänk för {scenario.title}"
                    >
                      {#if isSending[scenario.id]}
                        <span class="loading loading-spinner loading-xs"></span> Skapar...
                      {:else}
                        Skapa Länk
                      {/if}
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
  .container {
    max-width: 1200px; 
  }
  .table td, .table th {
    padding: 0.75rem; 
    vertical-align: top; 
  }
  .table td.whitespace-normal {
    white-space: normal; 
  }
  .badge-success {
    background-color: hsl(var(--su)); 
    color: hsl(var(--suc)); 
  }
  .badge-warning {
    background-color: hsl(var(--wa));
    color: hsl(var(--wac));
  }
  .badge-error {
    background-color: hsl(var(--er));
    color: hsl(var(--erc));
  }
</style>
