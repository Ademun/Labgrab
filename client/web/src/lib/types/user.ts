export interface User {
    username: string;
    name?: string;
    surname?: string;
    patronymic?: string;
    group_code?: string;
    phone_number?: string;
    photo_url?: string;
}

export interface UserDetails {
    name?: string;
    surname?: string;
    patronymic?: string;
    group_code?: string;
    phone_number?: string;
}