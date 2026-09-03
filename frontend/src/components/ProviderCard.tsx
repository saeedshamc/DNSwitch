import type { DNSProfile } from '../lib/types';
import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

function Star({ on }: { on: boolean }) {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" fill={on ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="1.8">
      <path d="M12 3.6l2.4 4.86 5.36.78-3.88 3.78.92 5.34L12 15.84 6.2 18.36l.92-5.34L3.24 9.24l5.36-.78z" />
    </svg>
  );
}

export default function ProviderCard({ profile }: { profile: DNSProfile }) {
  const lang = useAppStore((s) => s.lang);
  const ping = useAppStore((s) => s.pings[profile.id]);
  const fastest = useAppStore((s) => s.fastestId === profile.id);
  const favorite = useAppStore((s) => s.favorites.includes(profile.id));
  const applying = useAppStore((s) => s.applying);
  const testing = useAppStore((s) => s.testing);
  const applyProfile = useAppStore((s) => s.applyProfile);
  const testOne = useAppStore((s) => s.testOne);
  const toggleFavorite = useAppStore((s) => s.toggleFavorite);
  const openModal = useAppStore((s) => s.openModal);
  const i18n = t(lang);
  const title = lang === 'fa' && profile.nameFa ? profile.nameFa : profile.name;
  const servers = [...(profile.ipv4 || []), ...(profile.ipv6 || [])].filter(Boolean);

  return (
    <article
      className={`group relative flex flex-col rounded-3xl border bg-white p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md dark:bg-slate-900/80 ${
        fastest
          ? 'border-cyan-400 shadow-cyan-500/10 dark:border-cyan-400/70'
          : 'border-slate-200 dark:border-white/10'
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className="h-2.5 w-2.5 rounded-full" style={{ background: profile.color || '#64748b' }} />
          <h3 className="text-sm font-semibold">{title}</h3>
        </div>
        <div className="flex items-center gap-1">
          {fastest ? (
            <span className="rounded-full bg-cyan-500/15 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-cyan-600 dark:text-cyan-300">
              {i18n.fastest}
            </span>
          ) : null}
          <button
            type="button"
            className={`rounded-full p-1.5 ${favorite ? 'text-amber-400' : 'text-slate-400 hover:text-amber-300'}`}
            onClick={() => void toggleFavorite(profile.id)}
            aria-label={i18n.favorite}
          >
            <Star on={favorite} />
          </button>
        </div>
      </div>

      <p className="mt-3 min-h-[2.5rem] whitespace-pre-line font-mono text-xs leading-5 text-slate-500 dark:text-slate-400">
        {profile.isAutomatic ? i18n.dhcp : servers.join('\n') || i18n.noDns}
      </p>

      <div className="mt-auto flex items-center justify-between pt-3">
        <div className="text-xs text-slate-500">
          {profile.isAutomatic ? (
            <span>{i18n.reset}</span>
          ) : ping?.success ? (
            <span className="font-semibold text-emerald-500">
              {ping.latencyMs} {i18n.ms}
            </span>
          ) : ping?.error ? (
            <span className="text-rose-400">{i18n.timeout}</span>
          ) : (
            <span className="text-slate-400">—</span>
          )}
        </div>
        <div className="flex gap-1.5">
          {!profile.isPreset ? (
            <button type="button" className="btn-tiny" onClick={() => openModal(profile)}>
              {i18n.editProfile}
            </button>
          ) : null}
          {!profile.isAutomatic ? (
            <button type="button" className="btn-tiny" disabled={testing} onClick={() => void testOne(profile)}>
              {i18n.test}
            </button>
          ) : null}
          <button
            type="button"
            className="btn-tiny btn-tiny-primary"
            disabled={applying}
            onClick={() => void applyProfile(profile)}
          >
            {profile.isAutomatic ? i18n.reset : i18n.apply}
          </button>
        </div>
      </div>
    </article>
  );
}
