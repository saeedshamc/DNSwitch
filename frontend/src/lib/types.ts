export type Lang = 'en' | 'fa';
export type Theme = 'dark' | 'light';

export interface NetworkInterface {
  name: string;
  displayName: string;
  isUp: boolean;
  mtu: number;
  ipv4: string[];
  ipv6: string[];
  dns: string[];
  dhcp: boolean;
}

export interface DNSProfile {
  id: string;
  name: string;
  nameFa: string;
  ipv4: string[];
  ipv6: string[];
  isPreset: boolean;
  isAutomatic: boolean;
  color: string;
}

export interface AppSettings {
  language: string;
  theme: string;
  favorites: string[];
  customProfiles: DNSProfile[];
  lastInterface: string;
  applyToAll: boolean;
}

export interface ApplyResult {
  success: boolean;
  code: string;
  message: string;
  needsElevation: boolean;
}

export interface PingResult {
  profileId: string;
  server: string;
  latencyMs: number;
  success: boolean;
  error: string;
}

export type ToastKind = 'ok' | 'error' | 'info';

export interface ToastState {
  kind: ToastKind;
  text: string;
}
