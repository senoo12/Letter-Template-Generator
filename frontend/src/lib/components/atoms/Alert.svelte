<script lang="ts">
    import { fade } from 'svelte/transition';
    import { createEventDispatcher } from 'svelte';

    export let message: string | null = null;
    export let type: 'error' | 'success' | 'info' = 'error';

    const dispatch = createEventDispatcher();

    // Mapping warna berdasarkan type
    const styles = {
        error: 'bg-red-50 border-red-500 text-red-700',
        success: 'bg-green-50 border-green-500 text-green-700',
        info: 'bg-blue-50 border-blue-500 text-blue-700'
    };

    function close() {
        dispatch('close');
    }
</script>

{#if message}
    <div 
        transition:fade={{ duration: 200 }}
        class="fixed top-5 right-5 z-[100] w-80 p-4 rounded-xl border-l-4 shadow-lg {styles[type]}"
    >
        <div class="flex justify-between items-start">
            <div class="flex gap-3">
                <span class="font-bold">
                    {#if type === 'error'} ⚠️ {:else if type === 'success'} ✅ {:else} ℹ️ {/if}
                </span>
                <div>
                    <p class="font-bold text-sm capitalize">{type}</p>
                    <p class="text-sm opacity-90">{message}</p>
                </div>
            </div>
            <button on:click={close} class="text-lg leading-none hover:opacity-60">&times;</button>
        </div>
    </div>
{/if}