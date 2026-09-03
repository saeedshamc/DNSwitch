import { useAppStore } from '../store/useAppStore';

export default function Toast() {
  const toast = useAppStore((s) => s.toast);
  const showToast = useAppStore((s) => s.showToast);
  if (!toast) {
    return null;
  }
  const tone =
    toast.kind === 'ok'
      ? 'border-emerald-400/40 bg-emerald-500/15 text-emerald-100'
      : toast.kind === 'error'
        ? 'border-rose-400/40 bg-rose-500/15 text-rose-100'
        : 'border-cyan-400/40 bg-cyan-500/15 text-cyan-100';
  return (
    <button
      type="button"
      className={`fixed bottom-5 left-1/2 z-50 -translate-x-1/2 rounded-2xl border px-4 py-2 text-sm shadow-lg ${tone}`}
      onClick={() => showToast(null)}
    >
      {toast.text}
    </button>
  );
}
