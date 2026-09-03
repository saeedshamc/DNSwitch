import ProviderCard from '../components/ProviderCard';
import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function HomePage() {
  const lang = useAppStore((s) => s.lang);
  const presets = useAppStore((s) => s.presets);
  const customs = useAppStore((s) => s.customs);
  const openModal = useAppStore((s) => s.openModal);
  const i18n = t(lang);

  return (
    <main className="grid flex-1 grid-cols-1 gap-3 p-5 sm:grid-cols-2 xl:grid-cols-3">
      {presets.map((profile) => (
        <ProviderCard key={profile.id} profile={profile} />
      ))}
      {customs.map((profile) => (
        <ProviderCard key={profile.id} profile={profile} />
      ))}
      <button
        type="button"
        onClick={() => openModal(null)}
        className="flex min-h-[170px] items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-white/70 text-sm font-medium text-slate-500 transition hover:border-cyan-400 hover:text-cyan-600 dark:border-white/15 dark:bg-slate-900/40 dark:hover:border-cyan-400/70 dark:hover:text-cyan-300"
      >
        + {i18n.addCustom}
      </button>
    </main>
  );
}
