import { useState } from 'react';
import { t } from '../i18n';
import { useAppStore } from '../store/useAppStore';

export default function CustomProfileModal() {
  const lang = useAppStore((s) => s.lang);
  const editing = useAppStore((s) => s.editing);
  const closeModal = useAppStore((s) => s.closeModal);
  const saveProfile = useAppStore((s) => s.saveProfile);
  const deleteProfile = useAppStore((s) => s.deleteProfile);
  const i18n = t(lang);

  const [name, setName] = useState(editing?.name ?? '');
  const [v4a, setV4a] = useState(editing?.ipv4?.[0] ?? '');
  const [v4b, setV4b] = useState(editing?.ipv4?.[1] ?? '');
  const [v6a, setV6a] = useState(editing?.ipv6?.[0] ?? '');
  const [v6b, setV6b] = useState(editing?.ipv6?.[1] ?? '');

  if (!editing) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/50 p-4 backdrop-blur-sm">
      <form
        className="w-full max-w-lg rounded-3xl border border-slate-200 bg-white p-5 shadow-2xl dark:border-white/10 dark:bg-slate-900"
        onSubmit={(e) => {
          e.preventDefault();
          void saveProfile({
            ...editing,
            name,
            nameFa: name,
            ipv4: [v4a, v4b].map((s) => s.trim()).filter(Boolean),
            ipv6: [v6a, v6b].map((s) => s.trim()).filter(Boolean),
          });
        }}
      >
        <h2 className="text-lg font-semibold">{editing.id ? i18n.editProfile : i18n.addCustom}</h2>
        <div className="mt-4 grid gap-3">
          <label className="grid gap-1 text-sm">
            {i18n.name}
            <input className="field" value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="grid gap-1 text-sm">
              {i18n.ipv4Primary}
              <input className="field font-mono" value={v4a} onChange={(e) => setV4a(e.target.value)} />
            </label>
            <label className="grid gap-1 text-sm">
              {i18n.ipv4Secondary}
              <input className="field font-mono" value={v4b} onChange={(e) => setV4b(e.target.value)} />
            </label>
            <label className="grid gap-1 text-sm">
              {i18n.ipv6Primary}
              <input className="field font-mono" value={v6a} onChange={(e) => setV6a(e.target.value)} />
            </label>
            <label className="grid gap-1 text-sm">
              {i18n.ipv6Secondary}
              <input className="field font-mono" value={v6b} onChange={(e) => setV6b(e.target.value)} />
            </label>
          </div>
        </div>
        <div className="mt-5 flex items-center justify-between gap-2">
          {editing.id ? (
            <button
              type="button"
              className="btn-ghost text-rose-500"
              onClick={() => {
                if (window.confirm(i18n.confirmDelete)) {
                  void deleteProfile(editing.id);
                }
              }}
            >
              {i18n.deleteProfile}
            </button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <button type="button" className="btn-ghost" onClick={closeModal}>
              {i18n.cancel}
            </button>
            <button type="submit" className="btn-primary">
              {i18n.save}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
