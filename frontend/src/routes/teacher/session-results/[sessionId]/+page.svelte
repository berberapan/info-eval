<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { authStore, checkAuthClient } from '$lib/stores/auth'; // For auth checks

  // State variables
  let scenarioDetails = null;       // Full scenario structure (exercises, questions, options)
  let sessionDetails = null;        // Details of the specific scenario_session
  let allSessionResponses = [];   // Array of all student responses for this session
  
  let isLoading = true;
  let error = null;
  let currentScenarioSessionIdFromUrl = null;

  // API URLs - ensure these match your backend routes
  const SCENARIO_DETAIL_API_URL = 'http://localhost:9000/v1/scenario/';       // GET /v1/scenario/{scenario_id}
  const SESSION_DETAIL_API_URL = 'http://localhost:9000/v1/sessions/';         // GET /v1/sessions/{scenario_session_id}
  const SESSION_RESPONSES_API_URL = 'http://localhost:9000/v1/sessions/';    // GET /v1/sessions/{scenario_session_id}/responses

  async function fetchData(scenarioSessionId) {
    isLoading = true;
    error = null;
    scenarioDetails = null;
    sessionDetails = null;
    allSessionResponses = [];

    // Ensure user is authenticated to view teacher results
    await checkAuthClient();
    if (!$authStore.isAuthenticated) {
      error = "Åtkomst nekad. Du måste vara inloggad som lärare för att se resultat.";
      isLoading = false;
      // Optional: redirect to login
      // goto('/login?redirect=/teacher/session-results/' + scenarioSessionId);
      return;
    }

    try {
      // --- Step 1: Fetch ScenarioSession details (to get scenario_id and session notes/expiry) ---
      const sessionDetailRes = await fetch(`${SESSION_DETAIL_API_URL}${scenarioSessionId}`);
      if (!sessionDetailRes.ok) {
        const errData = await sessionDetailRes.json().catch(() => ({ error: `API Error: ${sessionDetailRes.status} - ${sessionDetailRes.statusText}` }));
        throw new Error(errData.error?.message || errData.error || `Failed to fetch session details: ${sessionDetailRes.statusText}`);
      }
      const fetchedSessionDetailContainer = await sessionDetailRes.json();
      // Assuming backend returns {"scenario_session": {...}} for GET /v1/sessions/:id
      sessionDetails = fetchedSessionDetailContainer.scenario_session; 

      if (!sessionDetails || !sessionDetails.id || !sessionDetails.scenario_id) {
        throw new Error("Invalid session details data or missing scenario_id.");
      }
      const scenarioId = sessionDetails.scenario_id;

      // --- Step 2: Fetch the full scenario structure using the scenario_id ---
      const scenarioRes = await fetch(`${SCENARIO_DETAIL_API_URL}${scenarioId}`);
      if (!scenarioRes.ok) {
        const errData = await scenarioRes.json().catch(() => ({ error: `API Error: ${scenarioRes.status} - ${scenarioRes.statusText}` }));
        throw new Error(errData.error?.message || errData.error || `Failed to fetch scenario structure: ${scenarioRes.statusText}`);
      }
      const fetchedScenarioContainer = await scenarioRes.json();
      // Assuming backend returns {"scenario": {...}} for GET /v1/scenario/:id
      scenarioDetails = fetchedScenarioContainer.scenario; 

      if (!scenarioDetails) {
        throw new Error("Scenario structure not found in API response.");
      }
      if (scenarioDetails.exercises && scenarioDetails.exercises.length > 0) {
          scenarioDetails.exercises.sort((a, b) => a.order - b.order);
      }

      // --- Step 3: Fetch all responses for this specific scenario_session_id ---
      // This endpoint GET /v1/sessions/{id}/responses should return {"session_responses": [...]}
      const responsesRes = await fetch(`${SESSION_RESPONSES_API_URL}${scenarioSessionId}/responses`);
      if (!responsesRes.ok) {
        const errData = await responsesRes.json().catch(() => ({ error: `API Error: ${responsesRes.status} - ${responsesRes.statusText}` }));
        throw new Error(errData.error?.message || errData.error || `Failed to fetch session responses: ${responsesRes.statusText}`);
      }
      const fetchedResponsesContainer = await responsesRes.json();
      // Assuming backend returns {"session_responses": [...]}
      allSessionResponses = fetchedResponsesContainer.session_responses || []; 

    } catch (e) {
      console.error("Error fetching teacher results data:", e);
      error = e.message || "An unknown error occurred while fetching results for this session.";
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    currentScenarioSessionIdFromUrl = $page.params.sessionId; // Matches the [sessionId] in the route
    if (currentScenarioSessionIdFromUrl) {
      fetchData(currentScenarioSessionIdFromUrl);
    } else {
      error = "Scenario Session ID is missing in the URL.";
      isLoading = false;
    }
  });

  // Helper to get all answers for a specific question from all responses
  function getAnswersForQuestion(questionId) {
    if (!allSessionResponses || allSessionResponses.length === 0) return [];
    return allSessionResponses.map(response => ({
      studentResponseId: response.id, // ID of the student's overall submission
      answer: response.raw_answers ? response.raw_answers[questionId] : undefined,
      aiFeedback: response.ai_feedback ? response.ai_feedback[questionId] : undefined,
      submittedAt: response.submitted_at
    })).filter(item => item.answer !== undefined); // Filter out if a student somehow didn't answer this specific question
  }
  
  // Helper to find a specific option by its ID from a list of options
  function getOptionById(options, optionId) {
    if (!options || !optionId) return null;
    return options.find(opt => opt.id === optionId);
  }

  // Helper to count occurrences of each option for multiple choice/true-false
  function getOptionCounts(question) {
      const answersForQ = getAnswersForQuestion(question.id);
      const counts = {};
      question.options.forEach(opt => counts[opt.id] = { text: opt.option_text, count: 0, is_correct: opt.is_correct });
      
      answersForQ.forEach(ans => {
          if (ans.answer && counts[ans.answer]) {
              counts[ans.answer].count++;
          }
      });
      return Object.values(counts); // Return as an array of objects
  }

</script>

<div class="container mx-auto p-4 md:p-8 min-h-screen bg-base-200 text-base-content">
  {#if isLoading}
    <div class="flex flex-col justify-center items-center h-96">
      <span class="loading loading-lg loading-spinner text-primary"></span>
      <p class="mt-4 text-xl">Laddar sessionsresultat...</p>
    </div>
  {:else if error}
    <div role="alert" class="alert alert-error">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>Fel: {error}</span>
      <div class="flex gap-2 mt-4">
        {#if error !== "Åtkomst nekad. Du måste vara inloggad som lärare för att se resultat."}
         <button on:click={() => fetchData(currentScenarioSessionIdFromUrl)} class="btn btn-sm btn-ghost">Försök igen</button>
        {/if}
        <button on:click={() => goto('/uppgifter')} class="btn btn-sm btn-outline">Tillbaka till scenarielistan</button>
      </div>
    </div>
  {:else if scenarioDetails && sessionDetails && allSessionResponses}
    <div class="mb-8 p-6 bg-base-100 rounded-lg shadow">
      <h1 class="text-3xl md:text-4xl font-bold mb-2">Resultatöversikt för Session</h1>
      <p class="text-xl">Scenario: <span class="font-semibold">{scenarioDetails.title}</span></p>
      {#if sessionDetails.notes}
        <p class="text-sm text-base-content/70 mt-1">Anteckningar för sessionen: {sessionDetails.notes}</p>
      {/if}
      <p class="text-sm text-base-content/70">Session skapad: {new Date(sessionDetails.created_at).toLocaleString()}</p>
      <p class="text-sm text-base-content/70">Sessionen går ut: {new Date(sessionDetails.expires_at).toLocaleString()}</p>
      <p class="text-sm font-medium mt-2">Totalt antal svar i denna session: {allSessionResponses.length}</p>
       <button class="btn btn-sm btn-outline mt-4" on:click={() => fetchData(currentScenarioSessionIdFromUrl)}>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
        </svg>
        Uppdatera resultat
      </button>
    </div>

    {#if allSessionResponses.length === 0}
        <div class="alert alert-info">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            <span>Inga svar har ännu skickats in för denna session.</span>
        </div>
    {/if}

    {#if scenarioDetails.exercises && scenarioDetails.exercises.length > 0}
      {#each scenarioDetails.exercises as exercise (exercise.id)}
        <div class="card bg-base-100 shadow-xl mb-6">
          <div class="card-body">
            <h3 class="card-title text-xl mb-2">
              Övning {exercise.order}: {#if exercise.title}{exercise.title}{:else}Frågor & Svar{/if}
            </h3>
            
            {#if exercise.questions && exercise.questions.length > 0}
              {#each exercise.questions as question (question.id)}
                {@const answersForThisQuestion = getAnswersForQuestion(question.id)}
                <div class="mb-6 p-4 border border-base-300 rounded-lg bg-base-200/30">
                  <p class="font-semibold text-lg mb-3">{question.question}</p>
                  
                  {#if question.type === 'true_false' || question.type === 'multiple_choice'}
                    {@const optionSummary = getOptionCounts(question)}
                    <h4 class="font-medium text-md mb-1">Svarsfördelning ({answersForThisQuestion.length} svar):</h4>
                    <ul class="list-disc list-inside space-y-1 text-sm">
                      {#each optionSummary as summary}
                        <li class="{summary.is_correct ? 'text-success font-semibold' : ''}">
                          {summary.text}: {summary.count} ({answersForThisQuestion.length > 0 ? ((summary.count / answersForThisQuestion.length) * 100).toFixed(1) : '0.0'}%)
                          {#if summary.is_correct} (Rätt svar){/if}
                        </li>
                      {/each}
                    </ul>
                  {:else if question.type === 'free_text'}
                    <h4 class="font-medium text-md mb-1">Inskickade fritextsvar ({answersForThisQuestion.length}):</h4>
                    {#if answersForThisQuestion.length > 0}
                      <div class="space-y-3 max-h-96 overflow-y-auto pr-2">
                        {#each answersForThisQuestion as individualAnswer (individualAnswer.studentResponseId)}
                          <div class="p-2 border border-base-300 rounded bg-base-100/50 text-sm">
                            <p class="whitespace-pre-wrap"><em>Svar:</em> {individualAnswer.answer || "Inget svar"}</p>
                            {#if individualAnswer.aiFeedback}
                              <p class="mt-1 pt-1 border-t border-base-300 whitespace-pre-wrap">
                                <span class="badge badge-info badge-xs mr-1 align-middle">AI</span> 
                                {individualAnswer.aiFeedback}
                              </p>
                            {:else}
                               <p class="mt-1 pt-1 border-t border-base-300 text-xs text-base-content/60">
                                (AI-återkoppling bearbetas eller saknas)
                               </p>
                            {/if}
                          </div>
                        {/each}
                      </div>
                    {:else}
                      <p class="text-sm text-base-content/70">Inga fritextsvar för denna fråga.</p>
                    {/if}
                  {/if}
                </div>
              {/each}
            {:else}
              <p class="text-sm text-base-content/70">Inga frågor i denna övning.</p>
            {/if}
          </div>
        </div>
      {/each}
    {:else}
        <p class="text-center text-lg p-4">Detta scenario innehåller inga övningar.</p>
    {/if}

    <div class="text-center mt-8">
        <button class="btn btn-primary" on:click={() => goto('/uppgifter')}>Tillbaka till Scenarielistan</button>
    </div>
  {:else if !isLoading} 
    <div class="text-center py-10 card bg-base-100 shadow-xl p-6">
      <p class="text-xl text-base-content mb-4">Kunde inte ladda all nödvändig data för resultatsidan. Kontrollera att sessionen och scenariot existerar.</p>
      <button on:click={() => goto('/uppgifter')} class="btn btn-primary">Tillbaka till Scenarielistan</button>
    </div>
  {/if}
</div>

<style>
  .container {
    max-width: 1000px; 
  }
   .whitespace-pre-wrap {
    white-space: pre-wrap; 
  }
</style>
