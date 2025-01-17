export interface User {
    ID?: string;
    id?: string;
    username: string;
    email: string;
    createdAt: string;
    updatedAt: string;
}

export interface Submission {
    ID?: string;
    id?: string;
    title: string;
    description: string;
    fileUrl: string;
    userId: string;
    packId: string;
    user?: User;
    createdAt: string;
    updatedAt: string;
}

export interface Sample {
    ID?: string;
    id?: string;
    title: string;
    description: string;
    fileUrl?: string;
    createdAt: string;
    updatedAt: string;
}

export interface SamplePack {
    ID?: string;
    id?: string;
    title: string;
    description: string;
    isActive: boolean;
    samples: Sample[];
    submissions: Submission[];
    uploadStart: string;
    uploadEnd: string;
    startDate: string;
    endDate: string;
    createdAt: string;
    updatedAt: string;
}
