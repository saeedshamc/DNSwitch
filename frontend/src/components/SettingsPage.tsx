import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function SettingsPage() {
  const lang = useAppStore((s) => s.lang);
  const theme = useAppStore((s) => s.theme);
  const platform = useAppStore((s) => s.platform);
  const elevated = useAppStore((s) => s.elevated);
  const logPath = useAppStore((s) => s.logPath);
  const setLang = useAppStore((s) => s.setLang);
  const setTheme = useAppStore((s) => s.setTheme);
  const openSettings = useAppStore((s) => s.openSettings);
  const requestElevation = useAppStore((s) => s.requestElevation);
  const quit = useAppStore((s) => s.quit);
  const i18n = t(lang);

  return (
    <div className="fixed inset-0 z-30 flex items-center justify-center bg-slate-950/50 p-4 backdrop-blur-sm">
      <section className="w-full max-w-xl rounded-3xl border border-slate-200 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-slate-900">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">{i18n.settings}</h2>
            <p className="mt-1 text-sm text-slate-500">{i18n.privacy}</p>
          </div>
          <button type="button" className="btn-ghost" onClick={() => openSettings(false)}>
            {i18n.close}
          </button>
        </div>

        <div className="mt-5 grid gap-4">
          <div>
            <div className="mb-2 text-xs font-medium uppercase tracking-[0.16em] text-slate-500">
              {i18n.language}
            </div>
            <div className="flex gap-2">
              <button type="button" className={`chip ${lang === 'en' ? 'chip-on' : ''}`} onClick={() => void setLang('en')}>
                {i18n.english}
              </button>
              <button type="button" className={`chip ${lang === 'fa' ? 'chip-on' : ''}`} onClick={() => void setLang('fa')}>
                {i18n.persian}
              </button>
            </div>
          </div>
          <div>
            <div className="mb-2 text-xs font-medium uppercase tracking-[0.16em] text-slate-500">{i18n.theme}</div>
            <div className="flex gap-2">
              <button type="button" className={`chip ${theme === 'dark' ? 'chip-on' : ''}`} onClick={() => void setTheme('dark')}>
                {i18n.dark}
              </button>
              <button type="button" className={`chip ${theme === 'light' ? 'chip-on' : ''}`} onClick={() => void setTheme('light')}>
                {i18n.light}
              </button>
            </div>
          </div>
          <div className="rounded-2xl border border-slate-200 p-4 text-sm dark:border-white/10">
            <div className="font-medium">{i18n.about}</div>
            <p className="mt-2 leading-6 text-slate-600 dark:text-slate-300">{i18n.aboutBody}</p>
            <p className="mt-3 text-xs text-slate-500">
              {i18n.version} · {platform || '—'} · {elevated ? i18n.elevated : i18n.notElevated}
            </p>
            <p className="mt-2 break-all font-mono text-[11px] text-slate-400">
              {i18n.logPath}: {logPath || '—'}
            </p>
            {!elevated ? (
              <button type="button" className="btn-primary mt-4" onClick={() => void requestElevation()}>
                {i18n.grantAccess}
              </button>
            ) : null}
            <button type="button" className="btn-ghost mt-3" onClick={() => quit()}>
              {i18n.quit}
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
