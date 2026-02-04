import {LabTopic, LabType} from "$lib/types/subscription.ts";

export function translateLabType(labType: string): string {
    switch (labType) {
        case LabType.Performance:
            return "Выполнение";
        case LabType.Defence:
            return "Защита";
        default:
            return labType;
    }
}

export function translateLabTopic(labTopic: string): string {
    switch (labTopic) {
        case LabTopic.Mechanics:
            return "Механика";
        case LabTopic.Electricity:
            return "Электричество";
        case LabTopic.Virtual:
            return "Виртуальная";
        case LabTopic.Optics:
            return "Оптика";
        case LabTopic.RigidBody:
            return "Твёрдое тело";
        default:
            return labTopic;
    }
}

export function translateTopicToEnglish(topic: string): string {
    switch (topic) {
        case "Механика":
            return LabTopic.Mechanics;
        case "Электричество":
            return LabTopic.Electricity;
        case "Виртуальная":
            return LabTopic.Virtual;
        case "Оптика":
            return LabTopic.Optics;
        case "Твёрдое тело":
            return LabTopic.RigidBody;
        default:
            return topic;
    }
}