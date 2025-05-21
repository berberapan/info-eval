<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let scenarioDetails = null; // To store the full scenario structure (exercises, questions, options)
  let sessionResponseData = null; // To store the specific student's response (raw_answers, ai_feedback)
  let isLoading = true;
  let error = null;
  let currentResponseIdFromUrl = null;

  const SCENARIO_DETAIL_API_URL = 'http://localhost:9000/v1/scenario/'; 
  const SESSION_DETAILS_API_URL = 'http://localhost:9000/v1/sessions/';
  const SESSION_RESPONSE_API_URL = 'http://localhost:9000/v1/session-responses/';

  async function fetchData(responseId) {
    isLoading = true;
    error = null;
    scenarioDetails = null;
    sessionResponseData = null;
    try {
      const resResponse = await fetch(`${SESSION_RESPONSE_API_URL}${responseId}`);
      if (!resResponse.ok) {
        const errData = await resResponse.json().catch(() => ({ error: `API Error: ${resResponse.status} - ${resResponse.statusText}` }));
        throw new Error(errData.error?.message || errData.error || `Failed to fetch session response: ${resResponse.statusText}`);
      }
      const fetchedSessionResponseContainer = await resResponse.json();
      // Backend returns {"session_response": {...}}
      sessionResponseData = fetchedSessionResponseContainer.session_response;
      if (!sessionResponseData || !sessionResponseData.id) {
        throw new Error("Invalid session response data or missing response ID.");
      }
      if (!sessionResponseData.scenario_session_id) {
          throw new Error("Missing scenario_session_id in session response data.");
      }
      const scenarioSessionDetailsRes = await fetch(`${SESSION_DETAILS_API_URL}${sessionResponseData.scenario_session_id}/scenario`);
      if (!scenarioSessionDetailsRes.ok) {
          const errData = await scenarioSessionDetailsRes.json().catch(() => ({ error: `API Error: ${scenarioSessionDetailsRes.status} - ${scenarioSessionDetailsRes.statusText}` }));
          throw new Error(errData.error?.message || errData.error || `Failed to fetch scenario session details: ${scenarioSessionDetailsRes.statusText}`);
      }
      const scenarioSessionDetails = await scenarioSessionDetailsRes.json();
      const scenarioId = scenarioSessionDetails.scenario_id; // Assuming backend returns { "scenario_id": "..." }
      if (!scenarioId) {
        throw new Error("Scenario ID not found in session details from API.");
      }
      const scenarioRes = await fetch(`${SCENARIO_DETAIL_API_URL}${scenarioId}`);
      if (!scenarioRes.ok) {
        const errData = await scenarioRes.json().catch(() => ({ error: `API Error: ${scenarioRes.status} - ${scenarioRes.statusText}` }));
        throw new Error(errData.error?.message || errData.error || `Failed to fetch scenario details: ${scenarioRes.statusText}`);
      }
      const fetchedScenarioContainer = await scenarioRes.json();
      scenarioDetails = fetchedScenarioContainer.scenario; // Assuming backend returns { "scenario": { ... } }
      if (!scenarioDetails) {
        throw new Error("Scenario details not found in API response.");
      }
       if (scenarioDetails.exercises && scenarioDetails.exercises.length > 0) {
          scenarioDetails.exercises.sort((a, b) => a.order - b.order);
      }
    } catch (e) {
      console.error("Error fetching results data:", e);
      error = e.message || "An unknown error occurred while fetching results.";
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    currentResponseIdFromUrl = $page.params.responseId;
    if (currentResponseIdFromUrl) {
      fetchData(currentResponseIdFromUrl);
    } else {
      error = "Response ID is missing in the URL.";
      isLoading = false;
    }
  });

  function getStudentAnswer(questionId) {
    return sessionResponseData?.raw_answers?.[questionId];
  }

  function getAIFeedback(questionId) {
    return sessionResponseData?.ai_feedback?.[questionId];
  }

  function getOptionById(options, optionId) {
    if (!options || !optionId) return null;
    return options.find(opt => opt.id === optionId);
  }

</script>

<div class="container mx-auto p-4 md:p-8 min-h-screen bg-base-200 text-base-content">
  {#if isLoading}
    <div class="flex flex-col justify-center items-center h-96">
      <span class="loading loading-lg loading-spinner text-primary"></span>
      <p class="mt-4 text-xl">Laddar resultat...</p>
    </div>
  {:else if error}
    <div role="alert" class="alert alert-error">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>Fel: {error}</span>
      <div class="flex gap-2 mt-4">
        <button on:click={() => fetchData(currentResponseIdFromUrl)} class="btn btn-sm btn-ghost">Försök igen</button>
        <button on:click={() => goto('/uppgifter')} class="btn btn-sm btn-outline">Tillbaka till uppgiftslistan</button>
      </div>
    </div>
  {:else if scenarioDetails && sessionResponseData}
    <div class="mb-8 p-6 bg-base-100 rounded-lg shadow">
      <h1 class="text-3xl md:text-4xl font-bold mb-2">Resultat för: {scenarioDetails.title}</h1>
      {#if scenarioDetails.description}
        <p class="text-lg text-base-content/80 mb-2">{scenarioDetails.description}</p>
      {/if}
      <p class="text-sm text-base-content/70">Svar skickades: {new Date(sessionResponseData.submitted_at).toLocaleString()}</p>
      <button class="btn btn-sm btn-outline mt-4" on:click={() => fetchData(currentResponseIdFromUrl)}>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
        </svg>
        Uppdatera AI-återkoppling
      </button>
    </div>

    {#if scenarioDetails.exercises && scenarioDetails.exercises.length > 0}
      {#each scenarioDetails.exercises as exercise (exercise.id)}
        <div class="card bg-base-100 shadow-xl mb-6">
          <div class="card-body">
            <h2 class="card-title text-2xl mb-1">
              Övning {exercise.order}: {#if exercise.title}{exercise.title}{:else}Resultat{/if}
            </h2>
            {#if exercise.info}
              <p class="mb-4 text-base-content/80 prose max-w-none">{exercise.info}</p>
            {/if}

            {#if exercise.questions && exercise.questions.length > 0}
              {#each exercise.questions as question (question.id)}
                {@const studentAnswerValue = getStudentAnswer(question.id)}
                {@const aiFeedbackText = getAIFeedback(question.id)}
                {@const selectedOptionDetails = (question.type === 'true_false' || question.type === 'multiple_choice') ? getOptionById(question.options, studentAnswerValue) : null}

                <div class="mb-6 p-4 border border-base-300 rounded-lg bg-base-200/30">
                  <p class="font-semibold text-lg mb-2">{question.question}</p>
                  
                  <div class="mb-2">
                    <span class="font-medium">Ditt svar:</span>
                    {#if question.type === 'free_text'}
                      <span class="italic ml-1 whitespace-pre-wrap">{studentAnswerValue || "Inget svar"}</span>
                    {:else if selectedOptionDetails}
                      <span class="italic ml-1">{selectedOptionDetails.option_text}</span>
                    {:else}
                      <span class="italic ml-1 text-base-content/60">Inget svar valt</span>
                    {/if}
                  </div>

                  <div class="mt-2 p-3 rounded-md text-sm 
                    { (question.type === 'true_false' || question.type === 'multiple_choice') ? 
                      (selectedOptionDetails && selectedOptionDetails.is_correct ? 'bg-success/20 text-success-content' : (studentAnswerValue ? 'bg-error/20 text-error-content' : 'bg-base-300/30')) :
                      (aiFeedbackText && !aiFeedbackText.toLowerCase().startsWith('error:') ? 'bg-info/20 text-info-content' : (aiFeedbackText ? 'bg-warning/20 text-warning-content' : (studentAnswerValue ? 'bg-base-300/30' : 'bg-base-300/30' )))
                    }">
                    <strong>Återkoppling:</strong>
                    {#if question.type === 'true_false' || question.type === 'multiple_choice'}
                      {#if selectedOptionDetails}
                        {selectedOptionDetails.feedback || (selectedOptionDetails.is_correct ? "Korrekt!" : "Felaktigt.")}
                      {:else if studentAnswerValue}
                        Ett fel uppstod vid visning av återkoppling för detta alternativ.
                      {:else}
                         Inget svar att ge återkoppling på.
                      {/if}
                    {:else if question.type === 'free_text'}
                      {#if aiFeedbackText}
                        <span class="whitespace-pre-wrap">{aiFeedbackText}</span>
                      {:else if studentAnswerValue}
                         AI-återkoppling bearbetas eller är inte tillgänglig ännu. Prova att uppdatera om en stund.
                      {:else}
                         Inget svar att ge återkoppling på.
                      {/if}
                    {/if}
                  </div>
                </div>
              {/each}
            {:else}
              <p>Inga frågor i denna övning.</p>
            {/if}
          </div>
        </div>
      {/each}
    {:else}
        <p class="text-center text-lg p-4">Detta scenario innehåller inga övningar.</p>
    {/if}
    <div class="text-center mt-8">
        <button class="btn btn-primary" on:click={() => goto('/uppgifter')}>Tillbaka till uppgiftslistan</button>
    </div>
  {:else}
    <div class="text-center py-10 card bg-base-100 shadow-xl p-6">
      <p class="text-xl text-base-content mb-4">Kunde inte ladda resultatdata. Kontrollera att ID är korrekt och försök igen.</p>
    </div>
  {/if}
</div>

<style>
  .container {
    max-width: 900px; /* Or your preferred max-width for results */
  }
  .whitespace-pre-wrap {
    white-space: pre-wrap; /* Ensures newlines in free text answers/feedback are respected */
  }
</style>
