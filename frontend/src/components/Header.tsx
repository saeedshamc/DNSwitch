import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function Header() {
  const lang = useAppStore((s) => s.lang);
  const status = useAppStore((s) => s.status);
  const testing = useAppStore((s) => s.testing);
  const applying = useAppStore((s) => s.applying);
  const elevated = useAppStore((s) => s.elevated);
  const flush = useAppStore((s) => s.flush);
  const testAll = useAppStore((s) => s.testAll);
  const openSettings = useAppStore((s) => s.openSettings);
  const i18n = t(lang);

  const label =
    testing || status === 'testing'
      ? i18n.statusTesting
      : applying
        ? i18n.applying
        : status === 'error'
          ? i18n.statusError
          : status === 'ok'
            ? i18n.statusOk
            : i18n.statusIdle;

  const dot =
    testing || applying
      ? 'bg-amber-400 animate-pulse'
      : status === 'error'
        ? 'bg-rose-400'
        : status === 'ok'
          ? 'bg-emerald-400'
          : 'bg-cyan-400';

  return (
    <header className="flex flex-wrap items-center gap-3 border-b border-slate-200/80 bg-white/80 px-5 py-3 backdrop-blur dark:border-white/10 dark:bg-slate-950/70">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-cyan-400 to-indigo-500 text-sm font-black text-slate-950 shadow-lg shadow-cyan-500/20">
          DNS
        </div>
        <div>
          <div className="text-base font-semibold tracking-tight">{i18n.appName}</div>
          <div className="text-xs text-slate-500 dark:text-slate-400">{i18n.tagline}</div>
        </div>
      </div>

      <div className="ms-auto flex flex-wrap items-center gap-2">
        <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs dark:border-white/10 dark:bg-white/5">
          <span className={`h-2 w-2 rounded-full ${dot}`} />
          <span>{label}</span>
          <span className="text-slate-400">·</span>
          <span className="text-slate-500 dark:text-slate-400">
            {elevated ? i18n.elevated : i18n.notElevated}
          </span>
        </div>
        <button type="button" className="btn-ghost" onClick={() => void testAll()} disabled={testing}>
          {testing ? i18n.testing : i18n.testAll}
        </button>
        <button type="button" className="btn-ghost" onClick={() => void flush()} disabled={applying}>
          {applying ? i18n.flushing : i18n.flush}
        </button>
        <button type="button" className="btn-ghost" onClick={() => openSettings(true)} aria-label={i18n.settings}>
          {i18n.settings}
        </button>
      </div>
    </header>
  );
}
