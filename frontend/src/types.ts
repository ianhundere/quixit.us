import type { SamplePack, Sample, User, Submission } from './types'

export interface SamplePack {
    ID: number;
    title: string;
    description: string;
    uploadStart: string;
    uploadEnd: string;
    startDate: string;
    endDate: string;
    isActive: boolean;
    samples: Sample[];
}

export interface Sample {
    ID: number;
    filename: string;
    fileUrl: string;
    user: User;
}

export interface User {
    ID: number;
    email: string;
}

export interface Submission {
    ID: number;
    title: string;
    description: string;
    fileUrl: string;
    user: User;
}
