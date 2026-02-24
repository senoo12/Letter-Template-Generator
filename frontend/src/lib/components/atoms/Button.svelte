<script lang="ts">
    export let type: "button" | "submit" = "button";
    export let disabled = false;
    export let isLoading = false;
</script>

<button
    {type}
    {disabled}
    class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-6 rounded-lg transition-all disabled:bg-gray-400 flex items-center justify-center gap-2"
    on:click
>
    {#if isLoading}
        <span class="animate-spin">⏳</span> 
    {/if}
    <slot />
</button>