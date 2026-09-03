import { create } from 'zustand';
import type {
  AppSettings,
  ApplyResult,
  DNSProfile,
  Lang,
  NetworkInterface,
  PingResult,
  Theme,
  ToastState,
} from '../lib/types';
import { detectLang, t } from '../i18n';
import {
  ApplyDNS,
  DeleteCustomProfile,
  FlushCache,
  GetInterfaces,
  GetLogPath,
  GetPlatform,
  GetPresets,
  GetSettings,
  IsElevated,
  QuitApplication,
  RequestElevation,
  ResetToDHCP,
  SaveCustomProfile,
  SetFavorite,
  SetPreferences,
  TestAll,
  TestProfile,
} from '../../wailsjs/go/main/App';

function resultMessage(lang: Lang, result: ApplyResult): string {
  const i18n = t(lang);
  const map: Record<string, string> = {
    applied: i18n.toastApplied,
    reset: i18n.toastReset,
    flushed: i18n.toastFlushed,
    saved: i18n.toastSaved,
    deleted: i18n.toastDeleted,
    settings: i18n.toastSaved,
    need_elevation: i18n.elevationNeeded,
    elevation_prompt: i18n.elevationBody,
    invalid_dns: i18n.toastInvalidDns,
    apply_failed: i18n.toastApplyFailed,
    invalid_interface: i18n.noInterfaces,
    invalid_profile: i18n.toastApplyFailed,
    config: i18n.toastApplyFailed,
    unsupported: i18n.toastApplyFailed,
  };
  return map[result.code] || result.message || i18n.statusError;
}

function asList<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

async function safeCall<T>(fn: () => Promise<T>, fallback: T): Promise<T> {
  try {
    const value = await fn();
    return value ?? fallback;
  } catch {
    return fallback;
  }
}

interface AppState {
  lang: Lang;
  theme: Theme;
  platform: string;
  elevated: boolean;
  logPath: string;
  interfaces: NetworkInterface[];
  selectedInterface: string;
  applyToAll: boolean;
  presets: DNSProfile[];
  customs: DNSProfile[];
  favorites: string[];
  pings: Record<string, PingResult>;
  fastestId: string;
  loading: boolean;
  applying: boolean;
  testing: boolean;
  settingsOpen: boolean;
  modalOpen: boolean;
  editing: DNSProfile | null;
  toast: ToastState | null;
  status: 'idle' | 'testing' | 'ok' | 'error';
  boot: () => Promise<void>;
  refreshInterfaces: () => Promise<void>;
  persistPrefs: (patch?: Partial<{ lang: Lang; theme: Theme; selectedInterface: string; applyToAll: boolean }>) => Promise<void>;
  setLang: (lang: Lang) => Promise<void>;
  setTheme: (theme: Theme) => Promise<void>;
  setInterface: (name: string) => Promise<void>;
  setApplyToAll: (value: boolean) => Promise<void>;
  applyProfile: (profile: DNSProfile) => Promise<void>;
  resetDhcp: () => Promise<void>;
  flush: () => Promise<void>;
  testOne: (profile: DNSProfile) => Promise<void>;
  testAll: () => Promise<void>;
  toggleFavorite: (id: string) => Promise<void>;
  saveProfile: (profile: DNSProfile) => Promise<void>;
  deleteProfile: (id: string) => Promise<void>;
  requestElevation: () => Promise<void>;
  quit: () => void;
  handleResult: (result: ApplyResult) => void;
  openSettings: (open: boolean) => void;
  openModal: (profile: DNSProfile | null) => void;
  closeModal: () => void;
  showToast: (toast: ToastState | null) => void;
}

function emptyProfile(): DNSProfile {
  return {
    id: '',
    name: '',
    nameFa: '',
    ipv4: ['', ''],
    ipv6: ['', ''],
    isPreset: false,
    isAutomatic: false,
    color: '#6366F1',
  };
}

export const useAppStore = create<AppState>((set, get) => ({
  lang: 'en',
  theme: 'dark',
  platform: '',
  elevated: false,
  logPath: '',
  interfaces: [],
  selectedInterface: '',
  applyToAll: false,
  presets: [],
  customs: [],
  favorites: [],
  pings: {},
  fastestId: '',
  loading: true,
  applying: false,
  testing: false,
  settingsOpen: false,
  modalOpen: false,
  editing: null,
  toast: null,
  status: 'idle',

  boot: async () => {
    try {
      const [settings, presets, ifaces, platform, elevated, logPath] = await Promise.all([
        safeCall(GetSettings, {} as AppSettings),
        safeCall(GetPresets, [] as DNSProfile[]),
        safeCall(GetInterfaces, [] as NetworkInterface[]),
        safeCall(GetPlatform, ''),
        safeCall(IsElevated, false),
        safeCall(GetLogPath, ''),
      ]);
      const ifaceList = asList(ifaces);
      const lang: Lang = settings.language === 'fa' || settings.language === 'en' ? settings.language : detectLang();
      const theme: Theme = settings.theme === 'light' ? 'light' : 'dark';
      const selected =
        settings.lastInterface && ifaceList.some((i) => i.name === settings.lastInterface)
          ? settings.lastInterface
          : ifaceList.find((i) => i.isUp)?.name || ifaceList[0]?.name || '';
      set({
        lang,
        theme,
        platform,
        elevated,
        logPath,
        presets: asList(presets),
        customs: asList(settings.customProfiles),
        favorites: asList(settings.favorites),
        applyToAll: Boolean(settings.applyToAll),
        interfaces: ifaceList,
        selectedInterface: selected,
        loading: false,
      });
      document.documentElement.lang = lang;
      document.documentElement.dir = lang === 'fa' ? 'rtl' : 'ltr';
      document.documentElement.classList.toggle('dark', theme === 'dark');
      if (!settings.language || !settings.lastInterface) {
        void SetPreferences(lang, theme, selected, Boolean(settings.applyToAll)).catch(() => undefined);
      }
    } catch {
      set({
        loading: false,
        toast: { kind: 'error', text: t(get().lang).toastLoadFailed },
      });
    }
  },

  refreshInterfaces: async () => {
    const ifaces = await safeCall(GetInterfaces, [] as NetworkInterface[]);
    set({ interfaces: asList(ifaces) });
  },

  persistPrefs: async (patch) => {
    const state = { ...get(), ...patch };
    if (patch) {
      set(patch);
    }
    document.documentElement.lang = state.lang;
    document.documentElement.dir = state.lang === 'fa' ? 'rtl' : 'ltr';
    document.documentElement.classList.toggle('dark', state.theme === 'dark');
    await SetPreferences(state.lang, state.theme, state.selectedInterface, state.applyToAll).catch(() => undefined);
  },

  setLang: async (lang) => {
    await get().persistPrefs({ lang });
  },

  setTheme: async (theme) => {
    await get().persistPrefs({ theme });
  },

  setInterface: async (name) => {
    await get().persistPrefs({ selectedInterface: name });
  },

  setApplyToAll: async (value) => {
    await get().persistPrefs({ applyToAll: value });
  },

  applyProfile: async (profile) => {
    set({ applying: true, status: 'idle' });
    try {
      let result: ApplyResult;
      if (profile.isAutomatic) {
        result = await ResetToDHCP(get().selectedInterface, get().applyToAll);
      } else {
        result = await ApplyDNS(
          get().selectedInterface,
          [...(profile.ipv4 || []), ...(profile.ipv6 || [])].filter(Boolean),
          get().applyToAll,
        );
      }
      get().handleResult(result);
      await get().refreshInterfaces();
    } catch {
      get().handleResult({ success: false, code: 'apply_failed', message: '', needsElevation: false });
    } finally {
      set({ applying: false });
    }
  },

  resetDhcp: async () => {
    set({ applying: true });
    try {
      const result = await ResetToDHCP(get().selectedInterface, get().applyToAll);
      get().handleResult(result);
      await get().refreshInterfaces();
    } catch {
      get().handleResult({ success: false, code: 'apply_failed', message: '', needsElevation: false });
    } finally {
      set({ applying: false });
    }
  },

  flush: async () => {
    set({ applying: true });
    try {
      get().handleResult(await FlushCache());
    } catch {
      get().handleResult({ success: false, code: 'apply_failed', message: '', needsElevation: false });
    } finally {
      set({ applying: false });
    }
  },

  testOne: async (profile) => {
    set({ testing: true, status: 'testing' });
    try {
      const ping = await TestProfile(profile);
      const pings = { ...get().pings, [profile.id]: ping };
      set({ pings, status: ping.success ? 'ok' : 'error' });
    } catch {
      set({ status: 'error' });
    } finally {
      set({ testing: false });
    }
  },

  testAll: async () => {
    set({ testing: true, status: 'testing' });
    try {
      const list = asList(await TestAll());
      const pings: Record<string, PingResult> = {};
      let fastestId = '';
      let best = Number.POSITIVE_INFINITY;
      for (const ping of list) {
        pings[ping.profileId] = ping;
        if (ping.success && ping.latencyMs > 0 && ping.latencyMs < best) {
          best = ping.latencyMs;
          fastestId = ping.profileId;
        }
      }
      set({ pings, fastestId, status: 'ok' });
    } catch {
      set({ status: 'error' });
    } finally {
      set({ testing: false });
    }
  },

  toggleFavorite: async (id) => {
    try {
      const favorite = !get().favorites.includes(id);
      await SetFavorite(id, favorite);
      const settings = await safeCall(GetSettings, {} as AppSettings);
      set({ favorites: asList(settings.favorites) });
    } catch {
      get().handleResult({ success: false, code: 'config', message: '', needsElevation: false });
    }
  },

  saveProfile: async (profile) => {
    try {
      const result = await SaveCustomProfile(profile);
      get().handleResult(result);
      if (result.success) {
        const settings = await safeCall(GetSettings, {} as AppSettings);
        set({ customs: asList(settings.customProfiles), modalOpen: false, editing: null });
      }
    } catch {
      get().handleResult({ success: false, code: 'apply_failed', message: '', needsElevation: false });
    }
  },

  deleteProfile: async (id) => {
    try {
      const result = await DeleteCustomProfile(id);
      get().handleResult(result);
      const settings = await safeCall(GetSettings, {} as AppSettings);
      set({
        customs: asList(settings.customProfiles),
        favorites: asList(settings.favorites),
        modalOpen: false,
        editing: null,
      });
    } catch {
      get().handleResult({ success: false, code: 'apply_failed', message: '', needsElevation: false });
    }
  },

  requestElevation: async () => {
    try {
      get().handleResult(await RequestElevation());
    } catch {
      get().handleResult({ success: false, code: 'need_elevation', message: '', needsElevation: true });
    }
  },

  quit: () => {
    void QuitApplication();
  },

  handleResult: (result) => {
    if (!result) {
      return;
    }
    const lang = get().lang;
    const text = resultMessage(lang, result);
    if (result.code === 'elevation_prompt' || (result.needsElevation && result.success)) {
      set({ status: 'idle', toast: { kind: 'info', text } });
      return;
    }
    if (!result.success || result.needsElevation) {
      set({ status: 'error', toast: { kind: 'error', text } });
      return;
    }
    set({
      status: 'ok',
      toast: { kind: 'ok', text: text || t(lang).toastSaved },
    });
  },

  openSettings: (open) => set({ settingsOpen: open }),
  openModal: (profile) => set({ modalOpen: true, editing: profile ?? emptyProfile() }),
  closeModal: () => set({ modalOpen: false, editing: null }),
  showToast: (toast) => set({ toast }),
}));
