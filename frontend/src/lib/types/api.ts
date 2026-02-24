export interface GenerateRequest {
    excel: File;
    template: File;
    base_file_field: string;
}

export interface GenerateResponse {
    success: boolean;
    message: string;
    data: {
        file_name: string;
        download_url: string;
    };
}

export interface ApiError {
    error: string;
}