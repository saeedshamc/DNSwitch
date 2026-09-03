import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function InterfaceBar() {
  const lang = useAppStore((s) => s.lang);
  const ifaces = useAppStore((s) => s.interfaces);
  const selected = useAppStore((s) => s.selectedInterface);
  const applyToAll = useAppStore((s) => s.applyToAll);
  const setInterface = useAppStore((s) => s.setInterface);
  const setApplyToAll = useAppStore((s) => s.setApplyToAll);
  const i18n = t(lang);

  const current = ifaces.find((i) => i.name === selected);
  const dnsText = current?.dhcp
    ? i18n.dhcp
    : current?.dns?.length
      ? current.dns.join(' · ')
      : i18n.noDns;

  return (
    <section className="mx-5 mt-5 rounded-3xl border border-slate-200 bg-white p-4 shadow-sm dark:border-white/10 dark:bg-slate-900/70">
      <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">
            {i18n.currentDns}
          </div>
          <div className="mt-1 font-mono text-sm text-slate-800 dark:text-slate-100">{dnsText}</div>
        </div>
        <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-600 dark:text-slate-300">
          <input
            type="checkbox"
            className="h-4 w-4 accent-cyan-500"
            checked={applyToAll}
            onChange={(e) => void setApplyToAll(e.target.checked)}
          />
          {i18n.applyToAll}
        </label>
      </div>
      <div className="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">{i18n.interfaces}</div>
      <div className="mt-2 flex flex-wrap gap-2">
        {ifaces.length === 0 ? (
          <p className="text-sm text-slate-500">{i18n.noInterfaces}</p>
        ) : (
          ifaces.map((iface) => {
            const active = iface.name === selected;
            return (
              <button
                key={iface.name}
                type="button"
                onClick={() => void setInterface(iface.name)}
                className={`rounded-2xl border px-3 py-2 text-start transition ${
                  active
                    ? 'border-cyan-400 bg-cyan-50 text-cyan-900 dark:border-cyan-400/60 dark:bg-cyan-400/10 dark:text-cyan-100'
                    : 'border-slate-200 bg-slate-50 hover:border-slate-300 dark:border-white/10 dark:bg-white/5 dark:hover:border-white/20'
                }`}
              >
                <div className="text-sm font-medium">{iface.displayName || iface.name}</div>
                <div className="mt-0.5 flex items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400">
                  <span className={iface.isUp ? 'text-emerald-500' : 'text-rose-400'}>
                    {iface.isUp ? i18n.up : i18n.down}
                  </span>
                  {iface.ipv4?.[0] ? <span className="font-mono">{iface.ipv4[0]}</span> : null}
                </div>
              </button>
            );
          })
        )}
      </div>
    </section>
  );
}
