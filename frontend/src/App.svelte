<script lang="ts">
  import { Link, Upload, ImagePlus, Zap } from 'lucide-svelte';

  let activeTab = 'url';
  let url = '';
  let productName = '';
  let cta = '';

  function handleGenerate() {
    console.log('Generating Ads...', {
      activeTab,
      url,
      productName,
      cta
    });
    alert(`Generating ads for ${productName || 'your product'}! Check console for details.`);
  }
</script>

<main class="min-h-screen bg-zinc-950 flex items-center justify-center p-4 md:p-8">
  <div class="w-full max-w-2xl bg-zinc-900/50 backdrop-blur-xl border border-white/10 rounded-3xl p-8 shadow-2xl relative overflow-hidden">
    <!-- Subtle Glow Effect -->
    <div class="absolute -top-24 -left-24 w-48 h-48 bg-purple-600/20 blur-[100px] rounded-full"></div>
    <div class="absolute -bottom-24 -right-24 w-48 h-48 bg-indigo-600/20 blur-[100px] rounded-full"></div>

    <div class="relative z-10 space-y-10">
      
      <!-- STEP 1: COMPETITOR RESEARCH -->
      <section class="space-y-6">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-violet-600 flex items-center justify-center text-white font-bold text-sm">
            1
          </div>
          <h2 class="text-xl font-semibold text-white tracking-tight">Competitor Research</h2>
        </div>

        <div class="space-y-4">
          <!-- Tabs UI -->
          <div class="flex p-1 bg-zinc-950 border border-white/5 rounded-xl">
            <button 
              class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg text-sm font-medium transition-all duration-200
                {activeTab === 'url' ? 'bg-zinc-800 text-white shadow-sm' : 'text-zinc-500 hover:text-zinc-300'}"
              on:click={() => activeTab = 'url'}
            >
              <Link size={16} />
              Competitor URL
            </button>
            <button 
              class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg text-sm font-medium transition-all duration-200
                {activeTab === 'upload' ? 'bg-zinc-800 text-white shadow-sm' : 'text-zinc-500 hover:text-zinc-300'}"
              on:click={() => activeTab = 'upload'}
            >
              <Upload size={16} />
              Upload References
            </button>
          </div>

          <!-- Conditional Content for Step 1 -->
          {#if activeTab === 'url'}
            <div class="relative group">
              <input 
                type="text" 
                bind:value={url}
                placeholder="Enter competitor's Instagram, Facebook, or Website URL..."
                class="w-full bg-zinc-950 border border-white/10 rounded-xl px-4 py-4 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-violet-500/50 focus:border-violet-500/50 transition-all"
              />
            </div>
          {:else}
            <div class="w-full border-2 border-dashed border-zinc-800 rounded-2xl bg-zinc-950/50 py-12 flex flex-col items-center justify-center gap-3 group hover:border-zinc-700 transition-colors cursor-pointer">
              <div class="p-3 rounded-full bg-zinc-900 group-hover:bg-zinc-800 transition-colors">
                <Upload class="text-zinc-500" size={24} />
              </div>
              <p class="text-zinc-500 text-sm font-medium">Drag & drop reference ad images here</p>
            </div>
          {/if}
        </div>
      </section>

      <!-- STEP 2: YOUR PRODUCT -->
      <section class="space-y-6">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-violet-600 flex items-center justify-center text-white font-bold text-sm">
            2
          </div>
          <h2 class="text-xl font-semibold text-white tracking-tight">Your Product</h2>
        </div>

        <div class="space-y-4">
          <!-- Drag & Drop Zone -->
          <div class="w-full border-2 border-dashed border-zinc-800 rounded-2xl bg-zinc-950/50 py-12 flex flex-col items-center justify-center gap-3 group hover:border-zinc-700 transition-colors cursor-pointer">
            <div class="p-3 rounded-full bg-zinc-900 group-hover:bg-zinc-800 transition-colors">
              <ImagePlus class="text-zinc-500" size={24} />
            </div>
            <p class="text-zinc-500 text-sm font-medium">Drag & drop your product image here</p>
          </div>

          <!-- Product Details Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <input 
              type="text" 
              bind:value={productName}
              placeholder="Product Name"
              class="bg-zinc-950 border border-white/10 rounded-xl px-4 py-4 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-violet-500/50 focus:border-violet-500/50 transition-all"
            />
            <input 
              type="text" 
              bind:value={cta}
              placeholder="Current Offer / CTA (e.g., 20% OFF today)"
              class="bg-zinc-950 border border-white/10 rounded-xl px-4 py-4 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-violet-500/50 focus:border-violet-500/50 transition-all"
            />
          </div>
        </div>
      </section>

      <!-- CTA BUTTON -->
      <button 
        on:click={handleGenerate}
        class="w-full bg-gradient-to-r from-violet-900 to-purple-800 hover:from-violet-800 hover:to-purple-700 text-white font-bold py-5 rounded-2xl flex items-center justify-center gap-3 shadow-[0_0_20px_rgba(139,92,246,0.3)] transition-all duration-300 hover:scale-[1.01] active:scale-[0.99] group"
      >
        <Zap size={20} class="fill-current group-hover:animate-pulse" />
        <span class="text-lg">Analyze Competitors & Generate Ads</span>
      </button>

    </div>
  </div>
</main>
