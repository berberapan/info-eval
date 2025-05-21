<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  // Reactive variables
  let scenarioData = null;
  let currentExerciseIndex = 0;
  let isLoading = true;
  let error = null;
  let selectedLeftTab = 0; 

  // Holds user's answers for ALL questions across ALL exercises
  let allScenarioAnswers = {}; // Changed from 'answers' to reflect it holds all answers

  // Submission state variables for the FINAL submission
  let isSubmittingFinalAnswers = false;
  let finalSubmissionError = null;
  let finalSubmissionSuccessMessage = null; // Will mainly be for navigation indication
  let lastSubmittedResponseId = null; 

  // API URLs
  const BASE_API_URL = 'http://localhost:9000/v1/scenario/'; 
  const SESSION_API_URL = 'http://localhost:9000/v1/sessions/'; 
  
  async function fetchScenarioData(sessionId) {
    isLoading = true;
    error = null;
    scenarioData = null;
    allScenarioAnswers = {}; // Reset all answers when fetching new scenario
    finalSubmissionError = null; 
    finalSubmissionSuccessMessage = null;
    lastSubmittedResponseId = null;

    try {
      const sessionResponse = await fetch(`${SESSION_API_URL}${sessionId}/scenario`);
      if (!sessionResponse.ok) {
        let errorMessage = sessionResponse.statusText;
        try {
          const errorData = await sessionResponse.json();
          errorMessage = errorData.message || errorData.detail || errorData.error || sessionResponse.statusText;
        } catch (jsonError) {
          try {
            const textError = await sessionResponse.text();
            if (textError && textError.length > 0 && textError.length < 300 && !textError.toLowerCase().includes('<html')) {
              errorMessage = textError;
            }
          } catch (textReadError) { /* ignore */ }
        }
        throw new Error(`Session API-fel! Status: ${sessionResponse.status} - ${errorMessage}`);
      }
      let sessionData;
      try {
        sessionData = await sessionResponse.json();
      } catch (e) {
        throw new Error("Kunde inte tolka sessionsvar från API. Förväntade JSON.");
      }
      const scenarioId = sessionData.scenario_id;
      if (!scenarioId) throw new Error("Scenario ID hittades inte i sessionsdata.");

      const scenarioResponse = await fetch(`${BASE_API_URL}${scenarioId}`);
      if (!scenarioResponse.ok) {
        let errorResponseMessage = scenarioResponse.statusText;
        try {
          const errorData = await scenarioResponse.json();
          errorResponseMessage = errorData.message || errorData.detail || errorData.error || scenarioResponse.statusText;
        } catch (jsonError) {
           try {
            const textError = await scenarioResponse.text();
            if (textError && textError.length > 0 && textError.length < 300 && !textError.toLowerCase().includes('<html')) {
              errorResponseMessage = textError;
            }
          } catch (textReadError) { /* ignore */ }
        }
        throw new Error(`Scenario API-fel! Status: ${scenarioResponse.status} - ${errorResponseMessage}`);
      }
      let data;
      try {
        data = await scenarioResponse.json();
      } catch (e) {
        throw new Error("Kunde inte tolka scenariosvar från API. Förväntade JSON.");
      }
      if (data && data.scenario) {
        scenarioData = data.scenario;
      } else if (data && !data.scenario) {
        scenarioData = data; 
      } else {
        throw new Error("Scenariodata är inte i det förväntade formatet eller saknas i API-svaret.");
      }
      if (scenarioData.exercises && scenarioData.exercises.length > 0) {
        scenarioData.exercises.sort((a, b) => a.order - b.order);
      }
      initializeAnswersForScenario(); // Initialize all answers once

    } catch (e) {
      error = e.message || "Ett okänt fel uppstod vid hämtning av scenario.";
      scenarioData = null;
    } finally {
      isLoading = false;
    }
  }

  // Initialize answers for ALL questions in the scenario if not already set
  function initializeAnswersForScenario() {
    if (scenarioData && scenarioData.exercises) {
      scenarioData.exercises.forEach(exercise => {
        if (exercise.questions) {
          exercise.questions.forEach(q => {
            // Only initialize if not already present (e.g., if user navigates back and forth)
            if (allScenarioAnswers[q.id] === undefined) {
                 allScenarioAnswers[q.id] = q.type === 'free_text' ? '' : null;
            }
            // Remove any local feedback properties, as feedback is on results page
            delete q.submittedFeedback; 
            delete q.wasCorrect;      
            delete q.aiFeedbackText; 
          });
        }
      });
    }
    // Ensure component updates if allScenarioAnswers was modified
    allScenarioAnswers = { ...allScenarioAnswers };
  }
  
  // This function is called when moving between exercises to reset tab, not answers
  function prepareCurrentExerciseDisplay() {
    selectedLeftTab = 0; 
    // No longer clearing submission messages here as submission is only at the end
  }


  $: currentExercise = scenarioData?.exercises?.[currentExerciseIndex];
  $: currentMedia = currentExercise?.media;
  $: totalTabs = 1 + (currentMedia?.length || 0);
  $: isLastExercise = scenarioData && scenarioData.exercises && currentExerciseIndex === scenarioData.exercises.length - 1;

  onMount(() => {
    const sessionId = $page.params.sessionId; 
    if (sessionId) {
      fetchScenarioData(sessionId);
    } else {
      error = "Scenario Session ID saknas i URLen.";
      isLoading = false;
    }
  });

  function navigateExercise(direction) {
    const numExercises = scenarioData?.exercises?.length || 0;
    let newIndex = currentExerciseIndex + direction;

    if (newIndex >= 0 && newIndex < numExercises) {
      currentExerciseIndex = newIndex;
      prepareCurrentExerciseDisplay(); // Reset tab, but don't re-initialize/clear answers
    }
  }

  function handleRadioAnswer(questionId, optionId) {
    allScenarioAnswers = { ...allScenarioAnswers, [questionId]: optionId };
  }

  function handleTextAnswer(questionId, event) {
    allScenarioAnswers = { ...allScenarioAnswers, [questionId]: event.target.value };
  }

  // This function is now only called on the LAST exercise
  async function submitFinalAnswers() {
    if (!isLastExercise || isSubmittingFinalAnswers) return;

    isSubmittingFinalAnswers = true;
    finalSubmissionError = null;
    finalSubmissionSuccessMessage = null;
   
    const scenarioSessionId = $page.params.sessionId; 
    if (!scenarioSessionId) {
      finalSubmissionError = "Session ID saknas. Kan inte skicka svar.";
      isSubmittingFinalAnswers = false;
      return;
    }

    // API endpoint for submitting answers
    const submitUrl = `${SESSION_API_URL}${scenarioSessionId}/responses`;

    try {
      const response = await fetch(submitUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        // Send ALL accumulated answers
        body: JSON.stringify({ raw_answers: allScenarioAnswers }), 
      });

      if (!response.ok) {
        let errorMessage = `API-fel vid skickande av svar! Status: ${response.status}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.error?.message || errorData.error || errorData.message || errorMessage;
        } catch (e) { /* ignore if error response is not JSON */ }
        throw new Error(errorMessage);
      }

      const responseData = await response.json();
      if (responseData && responseData.session_response && responseData.session_response.id) {
        lastSubmittedResponseId = responseData.session_response.id; 
        finalSubmissionSuccessMessage = "Alla svar har skickats! Omdirigerar till resultatsidan...";
        
        // Navigate to results page
        setTimeout(() => {
            goto(`/session/result/${lastSubmittedResponseId}`);
        }, 1500); 
      } else {
        throw new Error("Kunde inte hämta ID för det skickade svaret från servern.");
      }
      
    } catch (err) {
      console.error("Fel vid skickande av slutgiltiga svar:", err);
      finalSubmissionError = err.message || "Ett okänt fel uppstod när svaren skulle skickas.";
      lastSubmittedResponseId = null; 
    } finally {
      isSubmittingFinalAnswers = false;
    }
  }

  // Check if all questions IN THE ENTIRE SCENARIO have been answered
  // This is for enabling the final submit button
  $: allQuestionsInScenarioAnswered = scenarioData && scenarioData.exercises && scenarioData.exercises.every(exercise => 
    exercise.questions && exercise.questions.every(q => {
      const answer = allScenarioAnswers[q.id];
      if (q.type === 'free_text') {
        return typeof answer === 'string' && answer.trim() !== '';
      }
      return answer !== null && answer !== undefined;
    })
  );

  function handleImageError(event) {
    event.target.alt = 'Bild kunde inte laddas.';
    event.target.src = 'https://placehold.co/600x400/cccccc/ffffff?text=Bild+saknas';
    console.warn("Image failed to load:", event.target.currentSrc || event.target.src);
  }
</script>

<div class="container mx-auto p-4 md:p-8 min-h-screen bg-base-200 text-base-content">
  {#if isLoading}
    <div class="flex flex-col justify-center items-center h-96">
      <span class="loading loading-lg loading-spinner text-primary"></span>
      <p class="mt-4 text-xl">Laddar scenario...</p>
    </div>
  {:else if error}
    <div role="alert" class="alert alert-error">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>Fel: {error}</span>
      <div class="flex gap-2 mt-4">
        <button on:click={() => fetchScenarioData($page.params.sessionId)} class="btn btn-sm btn-ghost">Försök igen</button>
        <button on:click={() => goto('/uppgifter')} class="btn btn-sm btn-outline">Tillbaka till listan</button>
      </div>
    </div>
  {:else if scenarioData}
    <div class="mb-8 p-6 bg-base-100 rounded-lg shadow">
      <h1 class="text-3xl md:text-4xl font-bold mb-2">{scenarioData.title}</h1>
      {#if scenarioData.description}
        <p class="text-lg text-base-content/80">{scenarioData.description}</p>
      {/if}
      <div class="mt-2 text-sm text-base-content/60">
        Svårighetsgrad: {scenarioData.difficulty} | Skapad: {new Date(scenarioData.created_at).toLocaleDateString()}
      </div>
    </div>

    {#if currentExercise}
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-2xl mb-1">
            Övning {currentExercise.order} av {scenarioData.exercises.length}
          </h2>
          {#if currentExercise.title}
            <p class="text-lg font-semibold mb-4">{currentExercise.title}</p>
          {/if}

          <div class="flex flex-col md:flex-row gap-6 mt-2">
            <div class="w-full md:w-1/2">
              <div role="tablist" class="tabs tabs-lifted tabs-lg mb-2">
                <button
                  role="tab"
                  aria-label="Information"
                  class="tab {selectedLeftTab === 0 ? 'tab-active font-semibold !bg-base-100' : 'bg-base-200/50'}"
                  on:click={() => selectedLeftTab = 0}>
                  Information
                </button>
                {#if currentMedia && currentMedia.length > 0}
                  {#each currentMedia as mediaItem, i (mediaItem.id)}
                    <button
                      role="tab"
                      aria-label="Media {i+1}"
                      class="tab {selectedLeftTab === i+1 ? 'tab-active font-semibold !bg-base-100' : 'bg-base-200/50'}"
                      on:click={() => selectedLeftTab = i+1}>
                      Media {i+1}
                    </button>
                  {/each}
                {/if}
              </div>
              <div class="p-4 bg-base-100 rounded-b-lg min-h-72">
                {#if selectedLeftTab === 0}
                  <div class="prose max-w-none">
                    {#if currentExercise.exercise_description}
                      <p class="text-base-content/90">{currentExercise.exercise_description}</p>
                    {:else}
                      <p class="text-base-content/70">Ingen beskrivning tillgänglig för denna övning.</p>
                    {/if}
                    {#if currentExercise.info}
                      <div class="mt-4">
                        <h3 class="text-lg font-medium">Mer information</h3>
                        <p>{currentExercise.info}</p>
                      </div>
                    {/if}
                  </div>
                {:else if currentMedia && selectedLeftTab - 1 < currentMedia.length}
                  {@const mediaContent = currentMedia[selectedLeftTab - 1]}
                  {#if mediaContent.media_type === 'image'}
                    <img
                      src={mediaContent.media_url}
                      alt="Scenario media {selectedLeftTab}: {mediaContent.caption || 'Bild'}"
                      class="rounded-lg shadow-md w-full h-auto object-contain max-h-[70vh]"
                      on:error={handleImageError}
                    />
                    {#if mediaContent.caption}
                      <p class="mt-2 text-sm text-center text-base-content/70">{mediaContent.caption}</p>
                    {/if}
                  {:else if mediaContent.media_type === 'video'}
                    <video controls class="rounded-lg shadow-md w-full max-h-[70vh]">
                      <source src={mediaContent.media_url} type="video/mp4" />
                      <track kind="captions" />
                      Din webbläsare stödjer inte videoelementet.
                    </video>
                    {#if mediaContent.caption}
                      <p class="mt-2 text-sm text-center text-base-content/70">{mediaContent.caption}</p>
                    {/if}
                  {:else}
                    <div class="p-4 bg-warning/20 rounded-lg text-warning-content">
                      <p class="font-semibold">Okänd mediatyp: {mediaContent.media_type}</p>
                      <p class="text-sm">URL: {mediaContent.media_url}</p>
                    </div>
                  {/if}
                {/if}
              </div>
            </div>

            <div class="w-full md:w-1/2 md:sticky md:top-8">
              <div class="p-4 bg-base-100 rounded-lg">
                <h3 class="text-xl font-semibold mb-4">Frågor</h3>
                
                {#if isLastExercise && finalSubmissionSuccessMessage}
                   <div role="alert" class="alert alert-info mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                    <span>{finalSubmissionSuccessMessage}</span>
                  </div>
                {/if}
                {#if isLastExercise && finalSubmissionError}
                  <div role="alert" class="alert alert-error mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2 2m2-2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    <span>{finalSubmissionError}</span>
                  </div>
                {/if}
                
                {#if currentExercise.questions && currentExercise.questions.length > 0}
                  {#each currentExercise.questions as question (question.id)}
                    <div class="mb-6 p-4 border border-base-300 rounded-lg bg-base-200/30">
                      <p class="font-semibold text-lg mb-3">{question.question}</p>
                      {#if question.type === 'true_false' || question.type === 'multiple_choice'}
                        <div class="space-y-2">
                          {#each question.options as option (option.id)}
                            <label class="flex items-center p-3 rounded-lg border hover:border-primary transition-all cursor-pointer {allScenarioAnswers[question.id] === option.id ? 'bg-primary/20 border-primary shadow-md' : 'bg-base-100 border-base-300'}">
                              <input
                                type="radio"
                                name="question-{question.id}"
                                class="radio radio-primary mr-3 shrink-0"
                                value={option.id}
                                checked={allScenarioAnswers[question.id] === option.id}
                                on:change={() => handleRadioAnswer(question.id, option.id)}
                                disabled={isSubmittingFinalAnswers && isLastExercise}
                              />
                              <span class="flex-grow">{option.option_text}</span>
                            </label>
                            {/each}
                        </div>
                      {:else if question.type === 'free_text'}
                        <textarea
                          class="textarea textarea-bordered w-full"
                          rows="4"
                          placeholder="Skriv ditt svar här..."
                          bind:value={allScenarioAnswers[question.id]}
                          disabled={isSubmittingFinalAnswers && isLastExercise}
                        ></textarea>
                         {:else}
                        <p class="text-warning">Okänd frågetyp: {question.type}</p>
                      {/if}
                    </div>
                  {/each}
                {:else}
                  <p class="text-center text-lg p-4">Inga frågor för denna övning.</p>
                {/if}
              </div>
            </div>
          </div>

          <div class="card-actions justify-between mt-8 items-center">
            <button
              class="btn btn-outline"
              on:click={() => navigateExercise(-1)}
              disabled={currentExerciseIndex === 0 || (isSubmittingFinalAnswers && isLastExercise)}
            >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-1"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
              Föregående
            </button>

            {#if isLastExercise}
              <button
                class="btn btn-primary btn-wide"
                on:click={submitFinalAnswers}
                disabled={!allQuestionsInScenarioAnswered || isSubmittingFinalAnswers}
              >
                {#if isSubmittingFinalAnswers}
                  <span class="loading loading-spinner loading-sm mr-2"></span> Skickar...
                {:else}
                  Skicka Alla Svar & Visa Resultat
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 ml-1"><path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" /></svg>
                {/if}
              </button>
            {:else}
             <div></div> 
            {/if}

            <button
              class="btn btn-outline"
              on:click={() => navigateExercise(1)}
              disabled={isLastExercise || (isSubmittingFinalAnswers && isLastExercise)}
            >
              Nästa
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 ml-1"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" /></svg>
            </button>
          </div>
        </div>
      </div>
    {:else if scenarioData.exercises && scenarioData.exercises.length === 0}
      <div class="text-center py-10 card bg-base-100 shadow-xl p-6">
        <p class="text-xl text-base-content mb-4">Detta scenario har för närvarande inga övningar.</p>
        <button on:click={() => goto('/uppgifter')} class="btn btn-secondary">Tillbaka till listan</button>
      </div>
    {:else}
      <div class="text-center py-10 card bg-base-100 shadow-xl p-6">
        <p class="text-xl text-base-content mb-4">Kunde inte ladda aktuell övning. Kontrollera konsolen för fel.</p>
        <button on:click={() => goto('/uppgifter')} class="btn btn-secondary">Tillbaka till listan</button>
      </div>
    {/if}
  {:else}
    <div class="text-center py-10 card bg-base-100 shadow-xl p-6">
      <p class="text-xl text-base-content mb-4">Inget scenariodata hittades eller datat är i ett oväntat format.</p>
      <div class="flex gap-2 justify-center">
        <button on:click={() => fetchScenarioData($page.params.sessionId)} class="btn btn-ghost">Försök igen</button>
        <button on:click={() => goto('/uppgifter')} class="btn btn-primary">Tillbaka till listan</button>
      </div>
    </div>
  {/if}
</div>

<style>
  .container {
    max-width: 1200px;
  }
  .md\:sticky {
    align-self: flex-start;
  }
</style>
