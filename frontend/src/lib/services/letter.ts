import { PUBLIC_BASE_URL } from '$env/static/public';
import type { GenerateResponse, GenerateRequest } from '$lib/types/api';

export const letterService = {
    async generateLetter(payload: GenerateRequest): Promise<GenerateResponse> {
        const formData = new FormData();
        formData.append('excel', payload.excel);
        formData.append('template', payload.template);
        formData.append('base_file_field', payload.base_file_field);

        const response = await fetch(`${PUBLIC_BASE_URL}/generate`, {
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
        window.location.href = `${PUBLIC_BASE_URL}${downloadPath}`;
    }
};