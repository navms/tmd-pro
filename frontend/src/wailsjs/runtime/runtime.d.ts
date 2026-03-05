export function EventsOn(eventName: string, callback: (...data: any[]) => void): () => void;
export function EventsOff(...eventNames: string[]): void;
export function EventsOnce(eventName: string, callback: (...data: any[]) => void): void;
export function EventsEmit(eventName: string, ...data: any[]): void;
export function WindowMinimise(): void;
export function WindowMaximise(): void;
export function Quit(): void;
