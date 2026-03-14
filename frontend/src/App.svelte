<script lang="ts">
  import {
    Building2, Palette, MessageSquarePlus, Layout,
    Plus, Trash2, Languages, Upload, FileType,
    Zap, X, CheckCircle2, Link as LinkIcon, ChevronDown,
    ArrowRight, ChevronLeft, Sparkles, Globe, Loader2, Download
  } from 'lucide-svelte';
  import { onMount } from 'svelte';

  // --- SVELTE 5 STATE (RUNES) ---
  let currentPage = $state('home'); // 'home' | 'builder' | 'result'
  let lang = $state('UA'); // 'UA' | 'EN'
  
  let companyName = $state('');
  let industry = $state('');
  let businessDescription = $state('');
  let brandColors = $state(['#8b5cf6', '#6366f1', '#a855f7']);
  let adPrompt = $state('');
  let selectedAdLanguage = $state('English');
  let competitorLinks = $state(['']);
  let myBusinessAdLibraryLink = $state('');
  
  let activePicker = $state<number | null>(null);
  let isIndustryOpen = $state(false);
  let isLanguageOpen = $state(false);
  let isDragging = $state(false);
  let uploadedFiles = $state<File[]>([]);
  let fileError = $state('');

  let isGenerating = $state(false);
  let generateError = $state('');
  let generatedAd = $state<{image_url: string, summary: string} | null>(null);

  const europeanLanguages = [
    'English', 'Ukrainian', 'German', 'French', 'Italian', 'Spanish', 'Polish', 
    'Portuguese', 'Dutch', 'Greek', 'Swedish', 'Czech', 'Hungarian', 'Romanian', 
    'Bulgarian', 'Danish', 'Finnish', 'Slovak', 'Norwegian', 'Croatian', 
    'Lithuanian', 'Slovenian', 'Estonian', 'Latvian', 'Irish', 'Maltese',
    'Icelandic', 'Albanian', 'Serbian', 'Bosnian', 'Macedonian'
  ].sort();

  const baseColors = [
    '#EF4444', '#F97316', '#F59E0B', '#10B981', 
    '#3B82F6', '#6366F1', '#8B5CF6', '#EC4899',
    '#000000', '#4B5563', '#9CA3AF', '#FFFFFF'
  ];

  const t = {
    UA: {
      title: 'TrendAd Builder',
      home: {
        hero: 'Steal the Winning Formula. Generate High-Converting Ads in Seconds.',
        subhero: 'Наш AI аналізує топ-10 оголошень ваших конкурентів — витягуючи їхні стратегії освітлення, композиції та копірайтингу — а потім генерує унікальні креативи для вашого продукту.',
        cta: 'Почати створення реклами'
      },
      nav: { back: 'На головну' },
      sections: {
        business: 'Business Info',
        assets: 'Brand Assets',
        prompt: 'Ad Generation',
        competitors: 'Analysis of competitors and own ads'
      },
      fields: {
        companyName: 'Як називається ваша компанія?',
        companyPlaceholder: 'Наприклад: CoffeeLab',
        industry: 'Чим займається ваш бізнес?',
        industrySelect: 'Оберіть сферу',
        description: 'Коротко опишіть ваш бізнес',
        descriptionPlaceholder: 'Ми продаємо органічну каву з доставкою по Україні.',
        upload: 'Додайте логотип або елементи айдентики',
        uploadHelper: 'Завантажте логотип, бренд-елементи або фото продукту.',
        uploadFormats: 'Підтримувані формати: PNG, SVG.',
        colors: 'Які основні кольори вашого бренду?',
        colorsHelper: 'Оберіть базовий колір або введіть свій HEX-код.',
        adPrompt: 'Яку рекламу ви хочете створити?',
        adPromptPlaceholder: 'Створіть Instagram рекламу для кав\'ярні. Стиль: мінімалістичний. Основний меседж: "Свіжа кава щодня".',
        adPromptHelper: 'Опишіть стиль реклами, продукт, основний меседж або цільову аудиторію.',
        adLanguage: 'Мова реклами',
        adLanguageSelect: 'Оберіть мову',
        competitors: 'Додайте посилання на конкурентів',
        competitorsHelper: 'Додайте до 10 посилань на Meta Ad Library.',
        competitorLabel: 'Посилання на Meta Ad Library конкурента',
        myBusinessLabel: 'Посилання на Meta Ad Library вашого бізнесу'
      },
      industries: ['ecommerce', 'restaurant / cafe', 'beauty', 'education', 'local services', 'other'],
      cta: 'Згенерувати рекламу',
      generating: 'Генерація...'
    },
    EN: {
      title: 'TrendAd Builder',
      home: {
        hero: 'Steal the Winning Formula. Generate High-Converting Ads in Seconds.',
        subhero: 'Our AI analyzes your competitors\' top 10 ads—extracting their lighting, composition, and copy strategies—then generates custom creatives for your product.',
        cta: 'Start Building Your Ad'
      },
      nav: { back: 'Back to Home' },
      sections: {
        business: 'Business Info',
        assets: 'Brand Assets',
        prompt: 'Ad Generation',
        competitors: 'Analysis of competitors and own ads'
      },
      fields: {
        companyName: 'What is your company name?',
        companyPlaceholder: 'Example: CoffeeLab',
        industry: 'What does your business do?',
        industrySelect: 'Select industry',
        description: 'Briefly describe your business',
        descriptionPlaceholder: 'We sell organic coffee with delivery across Europe.',
        upload: 'Upload your logo or brand assets',
        uploadHelper: 'Upload your logo, brand elements, or product images.',
        uploadFormats: 'Supported formats: PNG, SVG.',
        colors: 'What are your brand colors?',
        colorsHelper: 'Choose a base color or enter your HEX code.',
        adPrompt: 'What kind of ad do you want to generate?',
        adPromptPlaceholder: 'Create an Instagram ad for a coffee shop. Style: minimalistic. Main message: "Fresh coffee every day".',
        adPromptHelper: 'Describe the ad style, product, main message, or target audience.',
        adLanguage: 'Ad Language',
        adLanguageSelect: 'Select language',
        competitors: 'Add competitor links',
        competitorsHelper: 'Add up to 10 links to Meta Ad Library.',
        competitorLabel: 'Competitor Meta Ad Library Link',
        myBusinessLabel: 'Your Business Meta Ad Library Link'
      },
      industries: ['ecommerce', 'restaurant / cafe', 'beauty', 'education', 'local services', 'other'],
      cta: 'Generate Ad',
      generating: 'Generating...'
    }
  };

  // --- DERIVED STATE ---
  const cur = $derived(t[lang as keyof typeof t]);

  // --- LOGIC FUNCTIONS ---
  function toggleLang() {
    lang = lang === 'UA' ? 'EN' : 'UA';
  }

  function validateFile(file: File) {
    const validTypes = ['image/png', 'image/svg+xml', 'image/jpeg', 'image/webp'];
    if (!validTypes.includes(file.type)) {
      fileError = lang === 'UA' ? 'Тільки PNG, JPEG, WEBP або SVG!' : 'Only PNG, JPEG, WEBP or SVG allowed!';
      return false;
    }
    fileError = '';
    return true;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    const files = Array.from(e.dataTransfer?.files || []);
    const validFiles = files.filter(validateFile);
    uploadedFiles = [...uploadedFiles, ...validFiles];
  }

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    const validFiles = files.filter(validateFile);
    uploadedFiles = [...uploadedFiles, ...validFiles];
    input.value = ''; // Reset input to allow re-uploading same file
  }

  function removeFile(index: number) {
    uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
  }

  function togglePicker(e: MouseEvent, index: number) {
    e.stopPropagation();
    isIndustryOpen = false;
    isLanguageOpen = false;
    activePicker = activePicker === index ? null : index;
  }

  function toggleIndustry(e: MouseEvent) {
    e.stopPropagation();
    activePicker = null;
    isLanguageOpen = false;
    isIndustryOpen = !isIndustryOpen;
  }

  function toggleLanguage(e: MouseEvent) {
    e.stopPropagation();
    activePicker = null;
    isIndustryOpen = false;
    isLanguageOpen = !isLanguageOpen;
  }

  function selectIndustry(val: string) { industry = val; isIndustryOpen = false; }
  function selectAdLanguage(val: string) { selectedAdLanguage = val; isLanguageOpen = false; }
  function selectBaseColor(index: number, color: string) { brandColors[index] = color; activePicker = null; }
  
  function addCompetitor() { if (competitorLinks.length < 10) competitorLinks = [...competitorLinks, '']; }
  function removeCompetitor(index: number) { competitorLinks = competitorLinks.filter((_, i) => i !== index); }
  
  async function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = error => reject(error);
    });
  }

  function extractPageId(url: string): string {
    try {
      const urlObj = new URL(url);
      const pageId = urlObj.searchParams.get('view_all_page_id');
      if (pageId) return pageId;
      const match = url.match(/view_all_page_id=(\d+)/);
      return match ? match[1] : '';
    } catch {
      const match = url.match(/view_all_page_id=(\d+)/);
      return match ? match[1] : '';
    }
  }

  async function handleGenerate() {
    if (isGenerating) return;
    
    isGenerating = true;
    generateError = '';
    
    try {
      let logoBase64 = '';
      if (uploadedFiles.length > 0) {
        logoBase64 = await fileToBase64(uploadedFiles[0]);
      }

      let pageId = '';
      for (const link of competitorLinks) {
        pageId = extractPageId(link);
        if (pageId) break;
      }
      
      if (!pageId && myBusinessAdLibraryLink) {
         pageId = extractPageId(myBusinessAdLibraryLink);
      }

      if (!pageId) {
         pageId = "default_page_id";
      }

      const payload = {
        user_context: `Company: ${companyName}, Industry: ${industry}, Language: ${selectedAdLanguage}`,
        page_id: pageId,
        brand_info: {
          logo_image: logoBase64,
          company_description: businessDescription,
          company_colors: brandColors,
          creative_prompt: adPrompt
        }
      };

      const res = await fetch('/api/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || `Server error: ${res.status}`);
      }

      generatedAd = await res.json();
      currentPage = 'result';
    } catch (e: any) {
      console.error('Generation failed:', e);
      generateError = e.message || 'Unknown error occurred.';
    } finally {
      isGenerating = false;
    }
  }

  onMount(() => {
    const handleOutsideClick = () => {
      activePicker = null;
      isIndustryOpen = false;
      isLanguageOpen = false;
    };
    window.addEventListener('click', handleOutsideClick);
    return () => window.removeEventListener('click', handleOutsideClick);
  });
</script>

<main class="min-h-screen bg-[#0A031A] text-zinc-200 font-sans selection:bg-purple-500/30 overflow-x-hidden">
  
  {#if currentPage === 'home'}
    <!-- HOME PAGE -->
    <div class="relative min-h-screen flex flex-col items-center justify-center px-4 py-20 text-center">
      <div class="absolute top-1/4 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-indigo-600/10 blur-[120px] rounded-full -z-10"></div>
      <div class="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-purple-600/10 blur-[100px] rounded-full -z-10"></div>

      <div class="max-w-4xl space-y-10 animate-in fade-in slide-in-from-bottom-10 duration-1000">
        <div class="flex items-center justify-center gap-3 mb-8">
          <div class="p-2.5 bg-gradient-to-br from-purple-600 to-indigo-600 rounded-2xl shadow-xl shadow-purple-500/20">
            <Zap size={32} class="text-white fill-current" />
          </div>
          <span class="text-3xl font-bold tracking-tighter text-white">{cur.title}</span>
        </div>

        <h1 class="text-5xl md:text-7xl font-extrabold tracking-tight text-white leading-[1.1]">
          Steal the <span class="bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400">Winning Formula</span>.<br/>
          Generate High-Converting <br class="hidden md:block" /> Ads in Seconds.
        </h1>

        <p class="text-lg md:text-xl text-zinc-400 leading-relaxed max-w-3xl mx-auto">
          {cur.home.subhero}
        </p>

        <div class="pt-10 flex flex-col items-center gap-6">
          <button onclick={() => currentPage = 'builder'} class="group relative flex items-center gap-3 px-10 py-5 bg-white text-black font-bold rounded-2xl transition-all hover:scale-105 active:scale-95 shadow-[0_0_40px_rgba(255,255,255,0.15)]">
            <Sparkles size={20} class="text-purple-600" />
            <span class="text-lg">{cur.home.cta}</span>
            <ArrowRight size={20} class="transition-transform group-hover:translate-x-1" />
          </button>
          
          <button onclick={toggleLang} class="flex items-center gap-2 text-zinc-500 hover:text-white transition-colors text-sm font-medium">
            <Languages size={16} />
            {lang === 'UA' ? 'English' : 'Українська'}
          </button>
        </div>
      </div>
    </div>

  {:else if currentPage === 'builder'}
    <!-- BUILDER PAGE -->
    <div class="max-w-4xl mx-auto py-12 px-4 space-y-8 animate-in fade-in duration-500">
      
      <header class="flex justify-between items-center mb-12">
        <button onclick={() => currentPage = 'home'} class="flex items-center gap-2 px-4 py-2 bg-zinc-900 border border-white/5 rounded-full text-sm font-medium hover:bg-zinc-800 transition-colors">
          <ChevronLeft size={16} />
          {cur.nav.back}
        </button>
        
        <div class="flex items-center gap-3">
          <div class="p-1.5 bg-gradient-to-br from-purple-600 to-indigo-600 rounded-lg shadow-lg"><Zap size={18} class="text-white fill-current" /></div>
          <span class="text-lg font-bold text-white tracking-tight hidden sm:block">{cur.title}</span>
        </div>

        <button onclick={toggleLang} class="relative z-[110] flex items-center gap-2 px-4 py-2 bg-zinc-900 border border-white/5 rounded-full text-sm font-medium hover:bg-zinc-800 transition-colors">
          <Languages size={16} />
          {lang === 'UA' ? 'English' : 'Українська'}
        </button>
      </header>

      <div class="space-y-12 pb-24">
        
        <!-- 1. BUSINESS INFORMATION -->
        <section class="bg-zinc-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 space-y-8 relative isolate group transition-all {isIndustryOpen ? 'z-50 shadow-2xl shadow-indigo-500/10' : 'z-0'}">
          <div class="absolute -z-10 -top-24 -right-24 w-48 h-48 bg-purple-600/5 blur-[80px] rounded-full group-hover:bg-purple-600/10 transition-colors"></div>
          <div class="flex items-center gap-4">
            <span class="flex items-center justify-center w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-400 font-bold border border-indigo-500/20">1</span>
            <h2 class="text-xl font-semibold text-white">{cur.sections.business}</h2>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.companyName}</label>
              <input type="text" bind:value={companyName} placeholder="{cur.fields.companyPlaceholder}" class="w-full bg-zinc-800/40 border border-white/10 rounded-xl px-4 py-3.5 focus:ring-2 focus:ring-indigo-500/50 outline-none transition-all placeholder:text-zinc-500 text-white" />
            </div>

            <div class="space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.industry}</label>
              <div class="relative">
                <button onclick={toggleIndustry} class="w-full flex items-center gap-3 bg-zinc-800/40 border border-white/10 rounded-xl px-4 py-3.5 hover:border-white/20 transition-all text-left text-white group/drop">
                  <Building2 size={18} class="text-zinc-500 group-hover/drop:text-indigo-400 transition-colors" />
                  <span class="flex-1 truncate {industry ? 'text-white' : 'text-zinc-500'}">{industry || cur.fields.industrySelect}</span>
                  <ChevronDown size={18} class="text-zinc-600 transition-transform {isIndustryOpen ? 'rotate-180' : ''}" />
                </button>
                {#if isIndustryOpen}
                  <div class="absolute z-[100] top-full mt-3 left-0 w-full bg-[#0F0724] border border-white/10 rounded-2xl overflow-hidden shadow-2xl animate-in fade-in slide-in-from-top-2 p-2" onmousedown={(e) => e.stopPropagation()}>
                    <div class="max-h-60 overflow-y-auto custom-scroll">
                      {#each cur.industries as ind}
                        <button onclick={() => selectIndustry(ind)} class="w-full flex items-center justify-between px-4 py-3 rounded-xl transition-all text-sm {industry === ind ? 'bg-indigo-500/10 text-white font-medium' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-white text-left'}">
                          {ind} {#if industry === ind}<CheckCircle2 size={14} class="text-indigo-400" />{/if}
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <div class="md:col-span-2 space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.description}</label>
              <textarea bind:value={businessDescription} placeholder="{cur.fields.descriptionPlaceholder}" rows="3" class="w-full bg-zinc-800/40 border border-white/10 rounded-xl px-4 py-3.5 focus:ring-2 focus:ring-indigo-500/50 outline-none transition-all placeholder:text-zinc-500 text-white resize-none"></textarea>
            </div>
          </div>
        </section>

        <!-- 2. BRAND ASSETS -->
        <section class="bg-zinc-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 space-y-8 relative group transition-all {activePicker !== null ? 'z-50 shadow-2xl shadow-indigo-500/10' : 'z-0'}">
          <div class="flex items-center gap-4">
            <span class="flex items-center justify-center w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-400 font-bold border border-indigo-500/20">2</span>
            <h2 class="text-xl font-semibold text-white">{cur.sections.assets}</h2>
          </div>

          <div class="space-y-10">
            <!-- DRAG & DROP ZONE -->
            <div class="space-y-3">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.upload}</label>
              <div role="button" tabindex="0" ondragover={(e) => { e.preventDefault(); isDragging = true; }} ondragleave={() => isDragging = false} ondrop={handleDrop} onclick={() => document.getElementById('file-input')?.click()} class="w-full border-2 border-dashed rounded-2xl bg-zinc-800/20 py-10 flex flex-col items-center justify-center gap-3 transition-all cursor-pointer {isDragging ? 'border-indigo-500 bg-indigo-500/5' : 'border-zinc-800 hover:border-zinc-700'}">
                <input id="file-input" type="file" accept=".png,.svg,.jpg,.jpeg,.webp" multiple hidden onchange={handleFileSelect} />
                <div class="p-4 rounded-full bg-zinc-800/60 transition-colors"><Upload class={isDragging ? 'text-indigo-400' : 'text-zinc-500'} size={28} /></div>
                <div class="text-center px-6"><p class="text-zinc-400 text-sm font-medium">{cur.fields.uploadHelper}</p><p class="text-zinc-600 text-xs mt-1">{cur.fields.uploadFormats}</p></div>
                {#if fileError}<p class="text-red-400 text-xs mt-2 font-medium">{fileError}</p>{/if}
              </div>

              <!-- UPLOADED FILES LIST -->
              {#if uploadedFiles.length > 0}
                <div class="grid grid-cols-1 gap-3 mt-4">
                  {#each uploadedFiles as file, i}
                    <div class="w-full bg-zinc-800/40 border border-white/5 rounded-2xl p-4 flex items-center gap-4 animate-in fade-in zoom-in-95 duration-200">
                      <div class="p-3 rounded-xl bg-indigo-500/10 text-indigo-400"><FileType size={24} /></div>
                      <div class="flex-1 min-w-0"><p class="text-sm font-medium text-white truncate">{file.name}</p><p class="text-xs text-zinc-600">{(file.size / 1024).toFixed(1)} KB</p></div>
                      <button onclick={(e) => { e.stopPropagation(); removeFile(i); }} class="p-2 hover:bg-zinc-900 rounded-lg text-zinc-500 hover:text-red-400 transition-colors"><X size={20} /></button>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>

            <div class="space-y-5">
              <div class="flex flex-col gap-1"><label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.colors}</label><span class="text-xs text-zinc-600 ml-1">{cur.fields.colorsHelper}</span></div>
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                {#each brandColors as color, i}
                  <div class="relative">
                    <button onclick={(e) => togglePicker(e, i)} class="w-full flex items-center gap-3 bg-zinc-800/40 border border-white/10 rounded-2xl p-3 hover:border-white/20 transition-all text-left shadow-lg group/btn">
                      <div class="w-10 h-10 rounded-xl border border-white/10 shadow-lg shrink-0 transition-transform group-hover/btn:scale-105" style="background-color: {color}; box-shadow: 0 0 15px {color}40"></div>
                      <div class="flex-1 min-w-0"><p class="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">Color {i + 1}</p><p class="text-sm font-mono text-white truncate">{color}</p></div>
                      <ChevronDown size={16} class="text-zinc-600 group-hover/btn:text-zinc-400 transition-colors {activePicker === i ? 'rotate-180' : ''}" />
                    </button>
                    {#if activePicker === i}
                      <div class="absolute z-[100] top-full mt-3 left-0 w-full min-w-[220px] bg-[#0F0724] border border-white/10 rounded-2xl p-4 shadow-2xl animate-in fade-in slide-in-from-top-2 duration-200" onmousedown={(e) => e.stopPropagation()}>
                        <div class="grid grid-cols-4 gap-2 mb-4">
                          {#each baseColors as baseColor}
                            <button onclick={() => selectBaseColor(i, baseColor)} class="w-full aspect-square rounded-lg border border-white/5 hover:scale-110 hover:z-10 transition-transform active:scale-95" style="background-color: {baseColor}" title={baseColor}></button>
                          {/each}
                        </div>
                        <div class="space-y-2">
                          <label class="text-[10px] text-zinc-500 font-bold uppercase ml-1">Hex Code & Manual</label>
                          <div class="flex gap-2">
                            <input type="text" bind:value={brandColors[i]} maxlength="7" class="flex-1 bg-zinc-900 border border-white/5 rounded-xl px-4 py-2.5 text-sm font-mono text-white outline-none" />
                            <div class="relative w-10 h-10 shrink-0">
                              <input type="color" bind:value={brandColors[i]} class="absolute inset-0 w-full h-full rounded-xl bg-transparent border-none cursor-pointer overflow-hidden [&::-webkit-color-swatch-wrapper]:p-0 [&::-webkit-color-swatch]:border-none shadow-lg" />
                              <div class="absolute inset-0 rounded-xl border border-white/10 pointer-events-none flex items-center justify-center"><Palette size={14} class="text-white mix-blend-difference opacity-50" /></div>
                            </div>
                          </div>
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
          </div>
        </section>

        <!-- 3. AD GENERATION PROMPT -->
        <section class="bg-zinc-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 space-y-8 relative isolate group transition-all {isLanguageOpen ? 'z-50' : 'z-0'}">
          <div class="absolute -z-10 -top-24 -right-24 w-48 h-48 bg-violet-600/5 blur-[80px] rounded-full group-hover:bg-violet-600/10 transition-colors"></div>
          <div class="flex items-center gap-4">
            <span class="flex items-center justify-center w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-400 font-bold border border-indigo-500/20">3</span>
            <h2 class="text-xl font-semibold text-white">{cur.sections.prompt}</h2>
          </div>
          <div class="space-y-6">
            <div class="space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.adPrompt}</label>
              <div class="relative">
                <textarea bind:value={adPrompt} placeholder="{cur.fields.adPromptPlaceholder}" rows="5" class="w-full bg-zinc-800/40 border border-white/10 rounded-2xl px-5 py-5 focus:ring-2 focus:ring-indigo-500/50 outline-none transition-all placeholder:text-zinc-500 text-white leading-relaxed"></textarea>
                <div class="mt-3 flex items-start gap-2 text-xs text-zinc-500 ml-1"><MessageSquarePlus size={14} class="mt-0.5 text-zinc-600" /><span>{cur.fields.adPromptHelper}</span></div>
              </div>
            </div>

            <!-- AD LANGUAGE DROPDOWN -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.adLanguage}</label>
              <div class="relative max-w-xs">
                <button onclick={toggleLanguage} class="w-full flex items-center gap-3 bg-zinc-800/40 border border-white/10 rounded-xl px-4 py-3 hover:border-white/20 transition-all text-left text-white group/lang">
                  <Globe size={18} class="text-zinc-500 group-hover/lang:text-indigo-400 transition-colors" />
                  <span class="flex-1 truncate">{selectedAdLanguage || cur.fields.adLanguageSelect}</span>
                  <ChevronDown size={18} class="text-zinc-600 transition-transform {isLanguageOpen ? 'rotate-180' : ''}" />
                </button>
                {#if isLanguageOpen}
                  <div class="absolute z-[100] top-full mt-3 left-0 w-full bg-[#0F0724] border border-white/10 rounded-2xl overflow-hidden shadow-2xl animate-in fade-in slide-in-from-top-2 p-2" onmousedown={(e) => e.stopPropagation()}>
                    <div class="max-h-60 overflow-y-auto custom-scroll">
                      {#each europeanLanguages as l}
                        <button onclick={() => selectAdLanguage(l)} class="w-full flex items-center justify-between px-4 py-2.5 rounded-xl transition-all text-sm {selectedAdLanguage === l ? 'bg-indigo-500/10 text-white font-medium' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-white text-left'}">
                          {l} {#if selectedAdLanguage === l}<CheckCircle2 size={14} class="text-indigo-400" />{/if}
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>
          </div>
        </section>

        <!-- 4. COMPETITOR ANALYSIS -->
        <section class="bg-zinc-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 space-y-8 relative isolate group">
          <div class="flex items-center gap-4">
            <span class="flex items-center justify-center w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-400 font-bold border border-indigo-500/20">4</span>
            <h2 class="text-xl font-semibold text-white">{cur.sections.competitors}</h2>
          </div>
          <div class="space-y-6">
            <div class="space-y-2">
              <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.myBusinessLabel}</label>
              <div class="relative flex-1">
                <div class="absolute inset-y-0 left-4 flex items-center pointer-events-none"><LinkIcon size={14} class="text-zinc-600" /></div>
                <input type="url" bind:value={myBusinessAdLibraryLink} placeholder="https://..." class="w-full bg-zinc-800/40 border border-white/10 rounded-xl pl-11 pr-4 py-3.5 focus:ring-2 focus:ring-indigo-500/50 outline-none transition-all placeholder:text-zinc-500 text-sm text-white" />
              </div>
            </div>
            <div class="space-y-4 pt-2 border-t border-white/5">
              <div class="space-y-2">
                <label class="text-sm font-medium text-zinc-400 ml-1">{cur.fields.competitorLabel}</label>
                <div class="space-y-3">
                  {#each competitorLinks as link, i}
                    <div class="flex gap-3 animate-in fade-in slide-in-from-left-2 duration-300">
                      <div class="relative flex-1">
                        <div class="absolute inset-y-0 left-4 flex items-center pointer-events-none"><LinkIcon size={14} class="text-zinc-600" /></div>
                        <input type="url" bind:value={competitorLinks[i]} placeholder="https://..." class="w-full bg-zinc-800/40 border border-white/10 rounded-xl pl-11 pr-4 py-3 focus:ring-2 focus:ring-indigo-500/50 outline-none transition-all placeholder:text-zinc-500 text-sm text-white" />
                      </div>
                      {#if competitorLinks.length > 1}
                        <button onclick={() => removeCompetitor(i)} class="p-3 text-zinc-500 hover:text-red-400 hover:bg-red-400/10 rounded-xl transition-all"><Trash2 size={20} /></button>
                      {/if}
                    </div>
                  {/each}
                </div>
              </div>
              <div class="flex items-center justify-between pt-2">
                <p class="text-xs text-zinc-600 ml-1">{cur.fields.competitorsHelper}</p>
                {#if competitorLinks.length < 10}
                  <button onclick={addCompetitor} class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-indigo-400 hover:text-indigo-300 hover:bg-indigo-500/10 rounded-lg transition-all"><Plus size={16} />{lang === 'UA' ? 'Додати ще' : 'Add more'}</button>
                {/if}
              </div>
            </div>
          </div>
        </section>

        {#if generateError}
          <div class="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm font-medium text-center animate-in fade-in">
            {generateError}
          </div>
        {/if}

        <!-- FINAL CTA -->
        <div class="pt-10 flex justify-center">
          <button 
            onclick={handleGenerate} 
            disabled={isGenerating}
            class="group px-16 py-5 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-2xl shadow-[0_0_25px_rgba(139,92,246,0.4)] transition-all hover:scale-[1.05] active:scale-[0.95] text-lg tracking-wide flex items-center gap-3"
          >
            {#if isGenerating}
              <Loader2 size={24} class="animate-spin" />
              <span>{cur.generating}</span>
            {:else}
              {cur.cta}
            {/if}
          </button>
        </div>
      </div>
    </div>
  
  {:else if currentPage === 'result'}
    <!-- RESULT PAGE -->
    <div class="max-w-4xl mx-auto py-12 px-4 space-y-8 animate-in fade-in slide-in-from-bottom-8 duration-500">
      <header class="flex justify-between items-center mb-12">
        <button onclick={() => currentPage = 'builder'} class="flex items-center gap-2 px-4 py-2 bg-zinc-900 border border-white/5 rounded-full text-sm font-medium hover:bg-zinc-800 transition-colors">
          <ChevronLeft size={16} />
          {lang === 'UA' ? 'Назад до редактора' : 'Back to Builder'}
        </button>
        
        <div class="flex items-center gap-3">
          <div class="p-1.5 bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg shadow-lg"><CheckCircle2 size={18} class="text-white" /></div>
          <span class="text-lg font-bold text-white tracking-tight hidden sm:block">Success!</span>
        </div>
      </header>

      <div class="bg-zinc-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 space-y-8 relative isolate">
        <div class="absolute -z-10 top-0 left-1/2 -translate-x-1/2 w-full h-full bg-indigo-600/5 blur-[100px] rounded-full"></div>
        
        <h2 class="text-3xl font-bold text-white text-center mb-8">
          {lang === 'UA' ? 'Ваш згенерований креатив' : 'Your Generated Creative'}
        </h2>

        {#if generatedAd}
          <div class="flex flex-col items-center gap-8">
            <div class="w-full max-w-2xl aspect-square bg-black/50 border border-white/10 rounded-2xl overflow-hidden shadow-2xl relative group">
              <img src={generatedAd.image_url} alt="Generated Ad" class="w-full h-full object-contain" />
              
              <div class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center backdrop-blur-sm">
                <a href={generatedAd.image_url} target="_blank" download class="flex items-center gap-2 px-6 py-3 bg-white text-black font-bold rounded-xl hover:scale-105 transition-transform">
                  <Download size={20} />
                  Download Creative
                </a>
              </div>
            </div>

            <div class="w-full max-w-3xl bg-zinc-800/40 border border-white/10 rounded-2xl p-6 space-y-4">
              <div class="flex items-center gap-3 mb-4">
                <div class="p-2 bg-indigo-500/10 rounded-lg"><Sparkles size={20} class="text-indigo-400" /></div>
                <h3 class="text-lg font-semibold text-white">{lang === 'UA' ? 'Копірайтинг та інсайти' : 'Copywriting & Insights'}</h3>
              </div>
              <div class="prose prose-invert max-w-none">
                <p class="text-zinc-300 leading-relaxed whitespace-pre-wrap">{generatedAd.summary}</p>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</main>

<style>
  :global(::-webkit-scrollbar) { width: 8px; }
  :global(::-webkit-scrollbar-track) { background: transparent; }
  :global(::-webkit-scrollbar-thumb) { background: #272727; border-radius: 10px; }
  :global(::-webkit-scrollbar-thumb:hover) { background: #3f3f46; }
  
  .custom-scroll::-webkit-scrollbar { width: 4px; }
  .custom-scroll::-webkit-scrollbar-thumb { background: #312e81; border-radius: 10px; }
</style>