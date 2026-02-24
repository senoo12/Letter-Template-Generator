<script lang="ts">
    import { letterService } from '$lib/services/letter';
    import Button from '$lib/components/atoms/Button.svelte';
    import Alert from '$lib/components/atoms/Alert.svelte';
    import { fade, fly } from 'svelte/transition';

    let excelFile: File | null = null;
    let templateFile: File | null = null;
    let baseFileField = ""; 
    let loading = false;
    let alertMsg: string | null = null;
    let alertType: 'error' | 'success' = 'error';

    async function handleProcess() {
        alertMsg = null;
        if (!excelFile || !templateFile || !baseFileField) {
            alertMsg = "Mohon lengkapi data dan file terlebih dahulu.";
            alertType = 'error';
            return;
        }

        loading = true;
        try {
            const response = await letterService.generateLetter({
                excel: excelFile,
                template: templateFile,
                base_file_field: baseFileField
            });
            
            if (response.success) {
                alertMsg = "Dokumen berhasil diproses. Mengunduh...";
                alertType = 'success';
                letterService.downloadFile(response.data.download_url);
            }
        } catch (e: any) {
            alertMsg = e.message;
            alertType = 'error';
        } finally {
            loading = false;
        }
    }
</script>

<Alert message={alertMsg} type={alertType} on:close={() => alertMsg = null} />

<header class="pt-16 pb-8 px-4 text-center">
    <div in:fly={{ y: -20, duration: 800 }} class="max-w-3xl mx-auto">
        <span class="bg-indigo-100 text-indigo-700 px-4 py-1.5 rounded-full text-sm font-semibold tracking-wide uppercase">
            Productivity Tool
        </span>
        <h1 class="text-4xl md:text-5xl font-extrabold text-slate-900 mt-6 tracking-tight">
            Generate Ratusan Surat <br/> 
            <span class="text-indigo-600">Dalam Hitungan Detik.</span>
        </h1>
        <p class="mt-6 text-lg text-slate-600 leading-relaxed">
            Otomatisasi dokumen Anda dengan integrasi Excel dan Template Word. 
            Cepat, aman, dan tanpa ribet.
        </p>
    </div>
</header>

<section class="max-w-5xl mx-auto px-4 py-12">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div class="flex flex-col items-center text-center space-y-3 p-6 rounded-2xl bg-white/50 border border-slate-100 shadow-sm">
            <div class="w-12 h-12 bg-indigo-600 text-white rounded-full flex items-center justify-center font-bold text-xl shadow-lg shadow-indigo-200">1</div>
            <h3 class="font-bold text-slate-800">Preparasi Template</h3>
            <p class="text-sm text-slate-500 leading-relaxed">
                Siapkan dokumen Word (.docx) dan gunakan placeholder <code class="bg-indigo-50 text-indigo-600 px-1 rounded font-mono">{"{{nama_field}}"}</code> untuk data yang akan diganti secara dinamis.
            </p>
        </div>

        <div class="flex flex-col items-center text-center space-y-3 p-6 rounded-2xl bg-white/50 border border-slate-100 shadow-sm">
            <div class="w-12 h-12 bg-indigo-600 text-white rounded-full flex items-center justify-center font-bold text-xl shadow-lg shadow-indigo-200">2</div>
            <h3 class="font-bold text-slate-800">Sinkronisasi Data</h3>
            <p class="text-sm text-slate-500 leading-relaxed">
                Unggah Template dan Excel. Pastikan nama kolom di Excel sesuai dengan placeholder di Word, lalu tentukan <strong>Identifier Kolom</strong> sebagai nama file.
            </p>
        </div>

        <div class="flex flex-col items-center text-center space-y-3 p-6 rounded-2xl bg-white/50 border border-slate-100 shadow-sm">
            <div class="w-12 h-12 bg-indigo-600 text-white rounded-full flex items-center justify-center font-bold text-xl shadow-lg shadow-indigo-200">3</div>
            <h3 class="font-bold text-slate-800">Eksekusi & Unduh</h3>
            <p class="text-sm text-slate-500 leading-relaxed">
                Klik tombol generate untuk memproses dokumen. Sistem akan mengonversi data Anda ke dalam arsip ZIP yang siap digunakan secara instan.
            </p>
        </div>
    </div>
</section>

<main class="max-w-5xl mx-auto px-4 pb-24 flex justify-center mt-12">
    
    <div in:fade={{ delay: 200 }} class="bg-white p-8 rounded-[2rem] shadow-xl shadow-slate-200/60 border border-slate-100 relative overflow-hidden w-full max-w-2xl">
        <div class="absolute top-0 right-0 w-32 h-32 bg-indigo-50 rounded-full -mr-16 -mt-16 z-0"></div>
        
        <div class="relative z-10 space-y-6">
            <div>
                <label class="block text-sm font-bold text-slate-700 mb-2">Identifier Kolom</label>
                <input 
                    type="text" 
                    bind:value={baseFileField}
                    placeholder="Contoh: Nama_Siswa"
                    class="w-full px-4 py-3 bg-slate-50 rounded-xl border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:bg-white outline-none transition-all placeholder:text-slate-400"
                />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                
                <div class="group relative bg-blue-50 border-2 border-dashed border-blue-200 p-5 rounded-2xl hover:border-blue-400 transition-colors">
                    <label class="block text-xs font-bold text-blue-700 mb-3 uppercase tracking-wider">Template Word (.docx)</label>
                    <div class="relative">
                        <input 
                            type="file" 
                            accept=".docx" 
                            on:change={(e) => templateFile = e.currentTarget.files?.[0] || null}
                            class="block w-full text-[10px] text-blue-500 file:mr-3 file:py-2 file:px-3 file:rounded-lg file:border-0 file:text-[10px] file:font-bold file:bg-blue-600 file:text-white hover:file:bg-blue-700 cursor-pointer"
                        />
                    </div>
                    {#if templateFile}
                        <p class="mt-2 text-[10px] text-blue-600 font-medium truncate italic" title={templateFile.name}>
                            ✓ {templateFile.name}
                        </p>
                    {/if}
                </div>

                <div class="group relative bg-emerald-50 border-2 border-dashed border-emerald-200 p-5 rounded-2xl hover:border-emerald-400 transition-colors">
                    <label class="block text-xs font-bold text-emerald-700 mb-3 uppercase tracking-wider">Data Excel (.xlsx)</label>
                    <div class="relative">
                        <input 
                            type="file" 
                            accept=".xlsx" 
                            on:change={(e) => excelFile = e.currentTarget.files?.[0] || null}
                            class="block w-full text-[10px] text-emerald-500 file:mr-3 file:py-2 file:px-3 file:rounded-lg file:border-0 file:text-[10px] file:font-bold file:bg-emerald-600 file:text-white hover:file:bg-emerald-700 cursor-pointer"
                        />
                    </div>
                    {#if excelFile}
                        <p class="mt-2 text-[10px] text-emerald-600 font-medium truncate italic" title={excelFile.name}>
                            ✓ {excelFile.name}
                        </p>
                    {/if}
                </div>

            </div>

            <div class="pt-2">
                <Button isLoading={loading} disabled={loading} on:click={handleProcess}>
                    {loading ? 'Sedang Memproses...' : 'Mulai Generate Surat'}
                </Button>
            </div>
            
            <p class="text-center text-[10px] text-slate-400 uppercase tracking-widest font-bold">
                Output dalam format .ZIP
            </p>
        </div>
    </div>
</main>