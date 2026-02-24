import type { GenerateResponse, GenerateRequest } from '$lib/types/api';

const BASE_URL = 'https://etter-emplate-enerator-senoo128874-g8k2vocq1porsdi3n.leapcell-async.dev';

export const letterService = {
    async generateLetter(payload: GenerateRequest): Promise<GenerateResponse> {
        const formData = new FormData();
        formData.append('excel', payload.excel);
        formData.append('template', payload.template);
        formData.append('base_file_field', payload.base_file_field);

        const response = await fetch(`${BASE_URL}/generate`, {
            method: 'POST',
            body: formData,
        });

        const result = await response.json();

        if (!response.ok || !result.success) {
            throw new Error(result.error || result.message || 'Gagal memproses surat');
        }

        return result; 
    },

    downloadFile(downloadPath: string) {
        window.location.href = `${BASE_URL}${downloadPath}`;
    }
};