import { useEffect } from 'react';
import Header from './components/Header';
import InterfaceBar from './components/InterfaceBar';
import CustomProfileModal from './components/CustomProfileModal';
import SettingsPage from './components/SettingsPage';
import Toast from './components/Toast';
import HomePage from './pages/HomePage';
import { useAppStore } from './store/useAppStore';
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime';
import type { ApplyResult } from './lib/types';

export default function App() {
  const boot = useAppStore((s) => s.boot);
  const loading = useAppStore((s) => s.loading);
  const modalOpen = useAppStore((s) => s.modalOpen);
  const settingsOpen = useAppStore((s) => s.settingsOpen);
  const handleResult = useAppStore((s) => s.handleResult);
  const refreshInterfaces = useAppStore((s) => s.refreshInterfaces);
  const toast = useAppStore((s) => s.toast);
  const showToast = useAppStore((s) => s.showToast);

  useEffect(() => {
    void boot();
    EventsOn('dns-applied', (...data: unknown[]) => {
      const payload = (Array.isArray(data) ? data[0] : data) as ApplyResult;
      if (payload && typeof payload === 'object') {
        handleResult(payload);
        void refreshInterfaces();
      }
    });
    return () => {
      EventsOff('dns-applied');
    };
  }, [boot, handleResult, refreshInterfaces]);

  useEffect(() => {
    if (!toast) {
      return;
    }
    const timer = window.setTimeout(() => showToast(null), 4200);
    return () => window.clearTimeout(timer);
  }, [toast, showToast]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100 text-slate-500 dark:bg-slate-950 dark:text-slate-400">
        DNSwitch
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900 dark:bg-[#070b14] dark:text-slate-100">
      <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(circle_at_top,rgba(34,211,238,0.12),transparent_42%)]" />
      <div className="relative mx-auto flex min-h-screen max-w-6xl flex-col">
        <Header />
        <InterfaceBar />
        <HomePage />
      </div>
      {modalOpen ? <CustomProfileModal /> : null}
      {settingsOpen ? <SettingsPage /> : null}
      <Toast />
    </div>
  );
}
