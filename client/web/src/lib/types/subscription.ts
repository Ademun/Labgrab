export enum LabType {
    Performance = "Performance",
    Defence = "Defence"
}

export enum LabTopic {
    Mechanics = "Mechanics",
    Electricity = "Electricity",
    Virtual = "Virtual",
    Optics = "Optics",
    RigidBody = "Rigid Body"
}

export const LAB_TYPE_COLORS = {
    [LabType.Defence]: {
        bg: "bg-blue-500",
        text: "text-blue-500",
        hex: "#4A90E2",
        hover: "hover:bg-blue-600"
    },
    [LabType.Performance]: {
        bg: "bg-primary", // Оранжевый #FF6B35
        text: "text-primary",
        hex: "#FF6B35",
        hover: "hover:bg-primary/90"
    }
} as const;

export function needsAuditorium(labType: string): boolean {
    return labType === LabType.Performance;
}

export interface Subscription {
    uuid: string;
    lab_type: string;
    lab_topic: string;
    lab_number: number;
    lab_auditorium?: number;
    created_at: string;
    closed_at?: string;
}

export interface NewSubscription {
    lab_type: string;
    lab_topic: string;
    lab_number: number;
    lab_auditorium?: number;
    created_at: number;
}