<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  // Reactive variables
  let scenarioData = null;
  let currentExerciseIndex = 0;
  let isLoading = true;
  let error = null;
  let selectedLeftTab = 0; // First tab (0) will be info, subsequent tabs will be media

  // Holds user's answers for the current exercise's questions
  let answers = {};
  
  let isSubmittingAnswers = false;
  let submissionError = null;
  let submissionSuccessMessage = null;
  let submittedResponseData = null;

  // API URLs
  const BASE_API_URL = 'http://localhost:9000/v1/scenario/';
  const SESSION_API_URL = 'http://localhost:9000/v1/sessions/';
  const SUBMIT_ANSWERS_BASE_URL = 'http://localhost:9000/v1/sessions/';

  async function fetchScenarioData(sessionId) {
    isLoading = true;
    error = null;
    scenarioData = null;
    answers = {};
    submissionError = null; 
    submissionSuccessMessage = null;
    submittedResponseData = null;

    try {
      // Step 1: Get the scenario ID from the session
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
          } catch (textReadError) {
            // Ignore if reading as text also fails
          }
        }
        throw new Error(`Session API-fel! Status: ${sessionResponse.status} - ${errorMessage}`);
      }

      let sessionData;
      try {
        sessionData = await sessionResponse.json();
      } catch (e) {
        console.error("Failed to parse session API response as JSON:", e);
        throw new Error("Kunde inte tolka sessionsvar från API. Förväntade JSON.");
      }

      // Extract scenario ID from session data
      const scenarioId = sessionData.scenario_id;
      
      if (!scenarioId) {
        throw new Error("Scenario ID hittades inte i sessionsdata.");
      }

      // Step 2: Get the scenario data using the scenario ID
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
          } catch (textReadError) {
            // Ignore if reading as text also fails
          }
        }
        throw new Error(`Scenario API-fel! Status: ${scenarioResponse.status} - ${errorResponseMessage}`);
      }

      // Try to parse the successful response as JSON
      let data;
      try {
        data = await scenarioResponse.json();
      } catch (e) {
        console.error("Failed to parse scenario API response as JSON:", e);
        throw new Error("Kunde inte tolka scenariosvar från API. Förväntade JSON.");
      }

      if (data && data.scenario) {
        scenarioData = data.scenario;
        if (scenarioData.exercises && scenarioData.exercises.length > 0) {
          scenarioData.exercises.sort((a, b) => a.order - b.order);
        }
        initializeAnswersForCurrentExercise();
      } else if (data && !data.scenario) {
        console.warn("API response structure might have changed. Expected 'data.scenario' but got direct data object. Attempting to use direct data.");
        scenarioData = data;
        if (scenarioData.exercises && scenarioData.exercises.length > 0) {
          scenarioData.exercises.sort((a, b) => a.order - b.order);
        }
        initializeAnswersForCurrentExercise();
      } else {
        throw new Error("Scenariodata är inte i det förväntade formatet eller saknas i API-svaret.");
      }
    } catch (e) {
      console.error("Misslyckades med att hämta eller bearbeta scenario/sessionsdata:", e);
      error = e.message || "Ett okänt fel uppstod vid hämtning av scenario.";
      scenarioData = null;
    } finally {
      isLoading = false;
    }
  }

  // Initialize or clear answers for the current exercise
  function initializeAnswersForCurrentExercise() {
    answers = {};
    const exercise = scenarioData?.exercises?.[currentExerciseIndex];
    if (exercise && exercise.questions) {
      exercise.questions.forEach(q => {
        delete q.submittedFeedback;
        delete q.wasCorrect;
        delete q.aiFeedbackText;
        answers[q.id] = q.type === 'free_text' ? '' : null;
      });
    }
    selectedLeftTab = 0; // Reset to info tab
    submissionError = null; // Clear submission errors when changing exercise
    submissionSuccessMessage = null; // Clear submission success when changing exercise
  }

  // Reactive declarations for current exercise and media
  $: currentExercise = scenarioData?.exercises?.[currentExerciseIndex];
  $: currentMedia = currentExercise?.media;
  
  // Calculate total number of tabs (1 for info + number of media items)
  $: totalTabs = 1 + (currentMedia?.length || 0);

  // Load data when component mounts (client-side)
  onMount(() => {
    const sessionId = $page.params.sessionId;
    if (sessionId) {
      fetchScenarioData(sessionId);
    } else {
      error = "Scenario ID saknas i URLen.";
      isLoading = false;
    }
  });

  // Handle changing exercises
  function navigateExercise(direction) {
    const numExercises = scenarioData?.exercises?.length || 0;
    let newIndex = currentExerciseIndex + direction;

    if (newIndex >= 0 && newIndex < numExercises) {
      currentExerciseIndex = newIndex;
      initializeAnswersForCurrentExercise();
    }
  }

  // Handle answer selection
  function handleRadioAnswer(questionId, optionId) {
    answers = { ...answers, [questionId]: optionId };
  }

  function handleTextAnswer(questionId, event) {
    answers = { ...answers, [questionId]: event.target.value };
  }

  // Submit answers for the current exercise
  async function submitAnswers() {
    if (!currentExercise || !currentExercise.questions || isSubmittingAnswers) return;

    isSubmittingAnswers = true;
    submissionError = null;
    submissionSuccessMessage = null;

    console.log("Förbereder att skicka svar för övning:", currentExercise.id, "Svar:", answers);

    const sessionId = $page.params.sessionId; // This is the scenario_session_id
    if (!sessionId) {
      submissionError = "Session ID saknas. Kan inte skicka svar.";
      isSubmittingAnswers = false;
      return;
    }
    try {
      const response = await fetch(`${SUBMIT_ANSWERS_BASE_URL}${sessionId}/responses`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ raw_answers: answers }), // Backend expects {"raw_answers": {...}}
      });

      if (!response.ok) {
        let errorMessage = `API-fel vid skickande av svar! Status: ${response.status}`;
        try {
          const errorData = await response.json();
          // Prefer more specific error messages if available
          errorMessage = errorData.error?.message || errorData.error || errorData.message || errorMessage;
        } catch (e) {
          // Could not parse error JSON, stick with status text
        }
        throw new Error(errorMessage);
      }

      const responseData = await response.json();
      // Store the response from the backend. This will contain the created session_response object.
      // It will initially have ai_feedback as null or empty.
      submittedResponseData = responseData.session_response; 
      submissionSuccessMessage = "Svaren har skickats! Återkoppling för flervalsfrågor visas direkt. AI-återkoppling för fritextfrågor genereras och kommer att visas senare.";
      console.log("Svar skickade, serversvar:", submittedResponseData);
    
    currentExercise.questions.forEach(q => {
      if (q.type === 'true_false' || q.type === 'multiple_choice') {
        const selectedOption = q.options?.find(opt => opt.id === answers[q.id]);
        if (selectedOption) {
          q.submittedFeedback = selectedOption.feedback;
          q.wasCorrect = selectedOption.is_correct;
        } else {
          q.submittedFeedback = "Du måste välja ett alternativ.";
          q.wasCorrect = false;
        }
      } else if (q.type === 'free_text') {
        // Basic feedback for free_text; real validation would be backend.
        if (answers[q.id] && answers[q.id].trim() !== '') {
            q.submittedFeedback = "Svar för fritext mottaget.";
            q.wasCorrect = null; // Or some client-side hint if possible
        } else {
            q.submittedFeedback = "Du måste skriva ett svar.";
            q.wasCorrect = false;
        }
      }
    });
    // Force Svelte to recognize changes within the currentExercise object
    currentExercise = { ...currentExercise };
    } catch (err) {
      console.error("Fel vid skickande av svar:", err);
      submissionError = err.message || "Ett okänt fel uppstod när svaren skulle skickas.";
    } finally {
      isSubmittingAnswers = false;
    }
  }

  // Computed property to check if all questions in the current exercise have been answered
  $: allAnswered = currentExercise && currentExercise.questions && currentExercise.questions.every(q => {
    const answer = answers[q.id];
    if (q.type === 'free_text') {
      return typeof answer === 'string' && answer.trim() !== '';
    }
    return answer !== null && answer !== undefined;
  });

  // Computed property to check if feedback has been shown for any question in the current exercise
  $: feedbackShownForExercise = currentExercise && currentExercise.questions && currentExercise.questions.some(q => q.submittedFeedback);

  // Function to handle image loading errors
  function handleImageError(event) {
    event.target.alt = 'Bild kunde inte laddas.'; // Set alt text for accessibility
    event.target.src = 'https://placehold.co/600x400/cccccc/ffffff?text=Bild+saknas'; // Set a placeholder image
    console.warn("Image failed to load:", event.target.currentSrc || event.target.src); // Log the original failed src
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
      <div class="flex gap-2">
        <button on:click={() => fetchScenarioData($page.params.sessionId)} class="btn btn-sm btn-ghost">Försök igen</button>
        <button on:click={() => goto('/uppgifter')} class="btn btn-sm btn-outline">Tillbaka till listan</button>
      </div>
    </div>
  {:else if scenarioData}
    <div class="mb-8 p-6 bg-base-100 rounded-lg shadow">
      <h1 class="text-3xl md:text-4xl font-bold mb-2">{scenarioData.title}</h1>
      <p class="text-lg text-base-content/80">{scenarioData.description}</p>
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
            <!-- LEFT SIDE: Info and Media Tabs -->
            <div class="w-full md:w-1/2">
              <!-- Tabs for Info and Media -->
              <div role="tablist" class="tabs tabs-lifted tabs-lg mb-2">
                <!-- Info Tab -->
                <button
                  role="tab"
                  aria-label="Information"
                  class="tab {selectedLeftTab === 0 ? 'tab-active font-semibold !bg-base-100' : 'bg-base-200/50'}"
                  on:click={() => selectedLeftTab = 0}>
                  Information
                </button>
                
                <!-- Media Tabs -->
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

              <!-- Tab Content -->
              <div class="p-4 bg-base-100 rounded-b-lg min-h-72">
                <!-- Info Tab Content -->
                {#if selectedLeftTab === 0}
                  <div class="prose max-w-none">
                    {#if currentExercise.exercise_description}
                      <p class="text-base-content/90">{currentExercise.exercise_description}</p>
                    {:else}
                      <p class="text-base-content/70">Ingen beskrivning tillgänglig för denna övning.</p>
                    {/if}
                    
                    <!-- Additional info can be added here -->
                    {#if currentExercise.info}
                      <div class="mt-4">
                        <h3 class="text-lg font-medium">Mer information</h3>
                        <p>{currentExercise.info}</p>
                      </div>
                    {/if}
                  </div>
                <!-- Media Tab Content -->
                {:else if currentMedia && selectedLeftTab - 1 < currentMedia.length}
                  {#if currentMedia[selectedLeftTab - 1].media_type === 'image'}
                    <img
                      src={currentMedia[selectedLeftTab - 1].media_url}
                      alt="Scenario media {selectedLeftTab}: {currentMedia[selectedLeftTab - 1].caption || 'Bild'}"
                      class="rounded-lg shadow-md w-full h-auto object-contain max-h-[70vh]"
                      on:error={handleImageError}
                    />
                    {#if currentMedia[selectedLeftTab - 1].caption}
                      <p class="mt-2 text-sm text-center text-base-content/70">{currentMedia[selectedLeftTab - 1].caption}</p>
                    {/if}
                  {:else if currentMedia[selectedLeftTab - 1].media_type === 'video'}
                    <video controls class="rounded-lg shadow-md w-full max-h-[70vh]">
                      <source src={currentMedia[selectedLeftTab - 1].media_url} type="video/mp4" />
                      <track kind="captions" />
                      Din webbläsare stödjer inte videoelementet. Försök med en annan webbläsare.
                    </video>
                    {#if currentMedia[selectedLeftTab - 1].caption}
                      <p class="mt-2 text-sm text-center text-base-content/70">{currentMedia[selectedLeftTab - 1].caption}</p>
                    {/if}
                  {:else}
                    <div class="p-4 bg-warning/20 rounded-lg text-warning-content">
                      <p class="font-semibold">Okänd mediatyp: {currentMedia[selectedLeftTab - 1].media_type}</p>
                      <p class="text-sm">URL: {currentMedia[selectedLeftTab - 1].media_url}</p>
                    </div>
                  {/if}
                {/if}
              </div>
            </div>

            <!-- RIGHT SIDE: Questions -->
            <div class="w-full md:w-1/2 md:sticky md:top-8">
              <div class="p-4 bg-base-100 rounded-lg">
                <h3 class="text-xl font-semibold mb-4">Frågor</h3>
                
                {#if currentExercise.questions && currentExercise.questions.length > 0}
                  {#each currentExercise.questions as question (question.id)}
                    <div class="mb-6 p-4 border border-base-300 rounded-lg bg-base-200/30">
                      <p class="font-semibold text-lg mb-3">{question.question}</p>
                      {#if question.type === 'true_false' || question.type === 'multiple_choice'}
                        <div class="space-y-2">
                          {#each question.options as option (option.id)}
                            <label class="flex items-center p-3 rounded-lg border hover:border-primary transition-all cursor-pointer {answers[question.id] === option.id ? 'bg-primary/20 border-primary shadow-md' : 'bg-base-100 border-base-300'} {question.submittedFeedback ? 'opacity-70 cursor-not-allowed' : ''}">
                              <input
                                type="radio"
                                name="question-{question.id}"
                                class="radio radio-primary mr-3 shrink-0"
                                value={option.id}
                                checked={answers[question.id] === option.id}
                                on:change={() => handleRadioAnswer(question.id, option.id)}
                                disabled={question.submittedFeedback}
                              />
                              <span class="flex-grow">{option.option_text}</span>
                            </label>
                            {#if question.submittedFeedback && answers[question.id] === option.id}
                              <div class="mt-2 p-3 rounded-md text-sm {option.is_correct ? 'bg-success/20 text-success-content' : 'bg-error/20 text-error-content'}">
                                <strong>Feedback:</strong> {option.feedback}
                              </div>
                            {/if}
                          {/each}
                        </div>
                      {:else if question.type === 'free_text'}
                        <textarea
                          class="textarea textarea-bordered w-full"
                          rows="4"
                          placeholder="Skriv ditt svar här..."
                          value={answers[question.id] || ''}
                          on:input={(e) => handleTextAnswer(question.id, e)}
                          disabled={question.submittedFeedback}
                        ></textarea>
                         {#if question.submittedFeedback}
                            <div class="mt-2 p-3 rounded-md text-sm bg-info/20 text-info-content">
                              <strong>Feedback:</strong> {question.submittedFeedback}
                            </div>
                        {/if}
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
              disabled={currentExerciseIndex === 0}
            >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-1"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
              Föregående
            </button>

            <button
              class="btn btn-primary btn-wide"
              on:click={submitAnswers}
              disabled={!allAnswered || feedbackShownForExercise}
            >
                {#if feedbackShownForExercise}Svar Bedömda{:else}Skicka Svar{/if}
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 ml-1"><path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" /></svg>
            </button>

            <button
              class="btn btn-outline"
              on:click={() => navigateExercise(1)}
              disabled={!scenarioData || !scenarioData.exercises || currentExerciseIndex === scenarioData.exercises.length - 1}
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
  /* Additional custom styles */
  .container {
    max-width: 1200px;
  }
  
  /* Ensure sticky positioning works as expected */
  .md\:sticky {
    align-self: flex-start;
  }
</style>
