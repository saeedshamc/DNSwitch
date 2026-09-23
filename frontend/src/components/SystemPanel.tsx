import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function SystemPanel() {
  const lang = useAppStore((s) => s.lang);
  const dnsEnabled = useAppStore((s) => s.dnsEnabled);
  const proxy = useAppStore((s) => s.proxy);
  const applying = useAppStore((s) => s.applying);
  const setDNSEnabled = useAppStore((s) => s.setDNSEnabled);
  const setProxyField = useAppStore((s) => s.setProxyField);
  const applyProxy = useAppStore((s) => s.applyProxy);
  const clearProxy = useAppStore((s) => s.clearProxy);
  const i18n = t(lang);

  return (
    <section className="mx-5 mt-4 grid gap-4 rounded-3xl border border-slate-200 bg-white p-4 shadow-sm dark:border-white/10 dark:bg-slate-900/70 lg:grid-cols-2">
      <div className="rounded-2xl border border-slate-200 p-4 dark:border-white/10">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">
              {i18n.dnsControl}
            </div>
            <h3 className="mt-1 text-base font-semibold">{i18n.dnsToggle}</h3>
            <p className="mt-1 text-sm text-slate-500">{i18n.dnsToggleHint}</p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={dnsEnabled}
            disabled={applying}
            onClick={() => void setDNSEnabled(!dnsEnabled)}
            className={`relative h-8 w-14 shrink-0 rounded-full transition ${
              dnsEnabled ? 'bg-cyan-500' : 'bg-slate-300 dark:bg-slate-700'
            }`}
          >
            <span
              className={`absolute top-1 start-1 h-6 w-6 rounded-full bg-white shadow transition-transform ${
                dnsEnabled ? 'translate-x-6 rtl:-translate-x-6' : 'translate-x-0'
              }`}
            />
          </button>
        </div>
        <p className="mt-3 text-xs text-slate-500">
          {dnsEnabled ? i18n.dnsOn : i18n.dnsOff}
        </p>
      </div>

      <div className="rounded-2xl border border-slate-200 p-4 dark:border-white/10">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">
              {i18n.proxyControl}
            </div>
            <h3 className="mt-1 text-base font-semibold">{i18n.proxyTitle}</h3>
            <p className="mt-1 text-sm text-slate-500">{i18n.proxyHint}</p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={proxy.enabled}
            disabled={applying}
            onClick={() => {
              if (proxy.enabled) {
                void clearProxy();
              } else {
                setProxyField('enabled', true);
              }
            }}
            className={`relative h-8 w-14 shrink-0 rounded-full transition ${
              proxy.enabled ? 'bg-indigo-500' : 'bg-slate-300 dark:bg-slate-700'
            }`}
          >
            <span
              className={`absolute top-1 start-1 h-6 w-6 rounded-full bg-white shadow transition-transform ${
                proxy.enabled ? 'translate-x-6 rtl:-translate-x-6' : 'translate-x-0'
              }`}
            />
          </button>
        </div>

        <div className="mt-4 grid gap-2 sm:grid-cols-2">
          <label className="grid gap-1 text-xs text-slate-500">
            {i18n.proxyHttp}
            <input
              className="field text-sm text-slate-800 dark:text-slate-100"
              placeholder="127.0.0.1:8080"
              value={proxy.http}
              disabled={applying}
              onChange={(e) => setProxyField('http', e.target.value)}
            />
          </label>
          <label className="grid gap-1 text-xs text-slate-500">
            {i18n.proxyHttps}
            <input
              className="field text-sm text-slate-800 dark:text-slate-100"
              placeholder="127.0.0.1:8080"
              value={proxy.https}
              disabled={applying}
              onChange={(e) => setProxyField('https', e.target.value)}
            />
          </label>
          <label className="grid gap-1 text-xs text-slate-500">
            {i18n.proxySocks}
            <input
              className="field text-sm text-slate-800 dark:text-slate-100"
              placeholder="127.0.0.1:1080"
              value={proxy.socks}
              disabled={applying}
              onChange={(e) => setProxyField('socks', e.target.value)}
            />
          </label>
          <label className="grid gap-1 text-xs text-slate-500">
            {i18n.proxyNoProxy}
            <input
              className="field text-sm text-slate-800 dark:text-slate-100"
              placeholder="localhost,127.0.0.1"
              value={proxy.noProxy}
              disabled={applying}
              onChange={(e) => setProxyField('noProxy', e.target.value)}
            />
          </label>
        </div>

        <div className="mt-4 flex flex-wrap gap-2">
          <button
            type="button"
            className="btn-primary"
            disabled={applying}
            onClick={() => void applyProxy()}
          >
            {applying ? i18n.applying : i18n.applyProxy}
          </button>
          <button
            type="button"
            className="btn-ghost"
            disabled={applying}
            onClick={() => void clearProxy()}
          >
            {i18n.clearProxy}
          </button>
        </div>
      </div>
    </section>
  );
}
