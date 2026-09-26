"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { Floor, Library, Room } from "@/lib/types";
import Button from "./Button";
import Card from "./Card";
import Chip from "./Chip";
import ErrorState from "./ErrorState";
import Skeleton from "./Skeleton";

// Seed amenities; anything already on a library/room is added on top so nothing is un-selectable.
const BASE_AMENITIES = ["strong_wifi", "weak_wifi", "many_outlets", "aircon", "quiet", "clean", "cold"];
const fieldClass =
  "min-h-11 rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none focus-visible:border-brand";

function useErr() {
  const { t } = useT();
  return (e: unknown) => (e instanceof ApiError ? (e.status === 409 ? t("adminLib.hasRooms") : e.message) : t("error.network"));
}

function Amenities({ options, value, onChange }: { options: string[]; value: string[]; onChange: (v: string[]) => void }) {
  const { t } = useT();
  return (
    <fieldset>
      <legend className="mb-2 text-sm">{t("home.amenities")}</legend>
      <div className="flex flex-wrap gap-2">
        {options.map((a) => (
          <Chip
            key={a}
            label={a}
            pressed={value.includes(a)}
            onClick={() => onChange(value.includes(a) ? value.filter((x) => x !== a) : [...value, a])}
          />
        ))}
      </div>
    </fieldset>
  );
}

// Inline "are you sure" instead of window.confirm. 409 (library still has rooms) is surfaced verbatim.
function DeleteButton({ name, onDelete }: { name: string; onDelete: () => Promise<void> }) {
  const { t } = useT();
  const err = useErr();
  const [asking, setAsking] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function go() {
    setBusy(true);
    setError(null);
    try {
      await onDelete();
    } catch (e) {
      setError(err(e));
      setAsking(false);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-2">
      {asking ? (
        <div role="group" aria-label={t("adminLib.deleteConfirm", { name })} className="flex flex-wrap items-center gap-2">
          <p className="text-sm">{t("adminLib.deleteConfirm", { name })}</p>
          <Button variant="secondary" disabled={busy} onClick={() => setAsking(false)} autoFocus>
            {t("action.keep")}
          </Button>
          <Button variant="danger" loading={busy} onClick={go}>
            {t("action.delete")}
          </Button>
        </div>
      ) : (
        <Button variant="secondary" aria-label={`${t("action.delete")} ${name}`} onClick={() => setAsking(true)}>
          {t("action.delete")}
        </Button>
      )}
      {error && (
        <p role="alert" className="text-sm text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}

function FormShell({
  onSubmit, onCancel, submitLabel, error, children,
}: {
  onSubmit: () => Promise<void>;
  onCancel: () => void;
  submitLabel: string;
  error: string | null;
  children: React.ReactNode;
}) {
  const { t } = useT();
  const [busy, setBusy] = useState(false);
  return (
    <form
      className="flex flex-col gap-3"
      onSubmit={async (e) => {
        e.preventDefault();
        setBusy(true);
        await onSubmit();
        setBusy(false);
      }}
    >
      {children}
      {error && (
        <p role="alert" className="text-sm text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
      <div className="flex gap-2">
        <Button type="button" variant="secondary" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit" loading={busy}>
          {busy ? t("action.saving") : submitLabel}
        </Button>
      </div>
    </form>
  );
}

const Field = ({ label, ...rest }: { label: string } & React.InputHTMLAttributes<HTMLInputElement>) => (
  <label className="flex flex-col gap-1 text-sm">
    {label}
    <input required className={fieldClass} {...rest} />
  </label>
);

function LibraryForm({ initial, options, onSaved, onCancel }: { initial?: Library; options: string[]; onSaved: () => void; onCancel: () => void }) {
  const { t } = useT();
  const err = useErr();
  const [f, setF] = useState({
    name_th: initial?.name_th ?? "",
    name_en: initial?.name_en ?? "",
    capacity: String(initial?.capacity ?? ""),
    amenities: initial?.amenities ?? [],
    floors: initial?.floors ?? ([] as Floor[]),
  });
  const [error, setError] = useState<string | null>(null);
  const setFloor = (i: number, fl: Floor) => setF({ ...f, floors: f.floors.map((x, j) => (j === i ? fl : x)) });

  async function save() {
    setError(null);
    try {
      const body = JSON.stringify({ ...f, capacity: Number(f.capacity) });
      await api(initial ? `/api/libraries/${initial.id}` : "/api/libraries", { method: initial ? "PUT" : "POST", body });
      onSaved();
    } catch (e) {
      setError(err(e));
    }
  }

  return (
    <FormShell onSubmit={save} onCancel={onCancel} error={error} submitLabel={t(initial ? "action.save" : "adminLib.create")}>
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label={t("adminLib.nameTh")} value={f.name_th} onChange={(e) => setF({ ...f, name_th: e.target.value })} />
        <Field label={t("adminLib.nameEn")} value={f.name_en} onChange={(e) => setF({ ...f, name_en: e.target.value })} />
        <Field label={t("adminLib.capacity")} type="number" min={0} value={f.capacity} onChange={(e) => setF({ ...f, capacity: e.target.value })} />
      </div>
      <Amenities options={options} value={f.amenities} onChange={(amenities) => setF({ ...f, amenities })} />
      <fieldset className="flex flex-col gap-3">
        <legend className="mb-2 text-sm">{t("library.floors")}</legend>
        {f.floors.map((fl, i) => (
          <div key={i} className="flex flex-col gap-2 rounded-lg border border-border bg-surface-muted p-3">
            <div className="flex items-end gap-2">
              <div className="flex-1">
                <Field label={t("adminLib.floorName", { n: i + 1 })} value={fl.name} onChange={(e) => setFloor(i, { ...fl, name: e.target.value })} />
              </div>
              <Button type="button" variant="secondary" onClick={() => setF({ ...f, floors: f.floors.filter((_, j) => j !== i) })}>
                {t("adminLib.removeFloor", { n: i + 1 })}
              </Button>
            </div>
            {fl.zones.map((z, k) => (
              <div key={k} className="flex flex-wrap items-end gap-2 pl-3">
                <Field
                  label={t("adminLib.zoneName", { n: k + 1 })}
                  value={z.name}
                  onChange={(e) => setFloor(i, { ...fl, zones: fl.zones.map((x, j) => (j === k ? { ...x, name: e.target.value } : x)) })}
                />
                <Field
                  label={t("adminLib.zoneCapacity", { n: k + 1 })}
                  type="number"
                  min={0}
                  className={`${fieldClass} w-28`}
                  value={z.capacity}
                  onChange={(e) => setFloor(i, { ...fl, zones: fl.zones.map((x, j) => (j === k ? { ...x, capacity: Number(e.target.value) } : x)) })}
                />
                <Button type="button" variant="secondary" onClick={() => setFloor(i, { ...fl, zones: fl.zones.filter((_, j) => j !== k) })}>
                  {t("adminLib.removeZone", { n: k + 1 })}
                </Button>
              </div>
            ))}
            <Button type="button" variant="ghost" className="self-start" onClick={() => setFloor(i, { ...fl, zones: [...fl.zones, { id: "", name: "", capacity: 0 }] })}>
              <span aria-hidden>+ </span>{t("adminLib.addZone")}
            </Button>
          </div>
        ))}
        <Button type="button" variant="ghost" className="self-start" onClick={() => setF({ ...f, floors: [...f.floors, { id: "", name: "", zones: [] }] })}>
          <span aria-hidden>+ </span>{t("adminLib.addFloor")}
        </Button>
      </fieldset>
    </FormShell>
  );
}

function RoomForm({ libraryId, initial, options, onSaved, onCancel }: { libraryId: string; initial?: Room; options: string[]; onSaved: () => void; onCancel: () => void }) {
  const { t } = useT();
  const err = useErr();
  const [f, setF] = useState({ name: initial?.name ?? "", capacity: String(initial?.capacity ?? ""), amenities: initial?.amenities ?? [] });
  const [error, setError] = useState<string | null>(null);

  async function save() {
    setError(null);
    try {
      const body = JSON.stringify({ ...f, capacity: Number(f.capacity), library_id: libraryId });
      await api(initial ? `/api/rooms/${initial.id}` : "/api/rooms", { method: initial ? "PUT" : "POST", body });
      onSaved();
    } catch (e) {
      setError(err(e));
    }
  }

  return (
    <FormShell onSubmit={save} onCancel={onCancel} error={error} submitLabel={t(initial ? "action.save" : "adminLib.createRoom")}>
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label={t("adminLib.roomName")} value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} />
        <Field label={t("adminLib.capacity")} type="number" min={0} value={f.capacity} onChange={(e) => setF({ ...f, capacity: e.target.value })} />
      </div>
      <Amenities options={options} value={f.amenities} onChange={(amenities) => setF({ ...f, amenities })} />
    </FormShell>
  );
}

function RoomsPanel({ library, options }: { library: Library; options: string[] }) {
  const { t } = useT();
  const [rooms, setRooms] = useState<Room[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [editing, setEditing] = useState<string | null>(null); // room id | "new"

  const load = () =>
    api<Room[]>(`/api/libraries/${library.id}/rooms`)
      .then((r) => { setRooms(r); setFailed(false); })
      .catch(() => setFailed(true));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => { load(); }, [library.id]);
  const saved = () => { setEditing(null); load(); };

  return (
    <div className="flex flex-col gap-3 border-t border-border bg-surface-muted/50 px-4 py-3">
      {failed && <ErrorState onRetry={load} />}
      {rooms === null && !failed && <Skeleton className="h-10 w-full" />}
      {rooms?.length === 0 && editing !== "new" && <p className="text-sm text-muted-foreground">{t("adminLib.noRooms")}</p>}
      {rooms?.map((r) =>
        editing === r.id ? (
          <Card key={r.id}>
            <RoomForm libraryId={library.id} initial={r} options={options} onSaved={saved} onCancel={() => setEditing(null)} />
          </Card>
        ) : (
          <div key={r.id} className="flex flex-wrap items-start justify-between gap-3">
            <div className="min-w-0 text-sm">
              <p className="font-medium">{r.name}</p>
              <p className="text-xs text-muted-foreground">
                {t("adminLib.seats", { n: r.capacity })}
                {r.amenities.length > 0 && ` · ${r.amenities.join(" · ")}`}
              </p>
            </div>
            <div className="flex flex-wrap items-start gap-2">
              <Button variant="secondary" aria-label={`${t("action.edit")} ${r.name}`} onClick={() => setEditing(r.id)}>
                {t("action.edit")}
              </Button>
              <DeleteButton name={r.name} onDelete={() => api(`/api/rooms/${r.id}`, { method: "DELETE" }).then(load)} />
            </div>
          </div>
        ),
      )}
      {editing === "new" ? (
        <Card>
          <RoomForm libraryId={library.id} options={options} onSaved={saved} onCancel={() => setEditing(null)} />
        </Card>
      ) : (
        <Button variant="secondary" className="self-start" onClick={() => setEditing("new")}>
          <span aria-hidden>+ </span>{t("adminLib.addRoom")}
        </Button>
      )}
    </div>
  );
}

export default function AdminLibraries() {
  const { t, lang } = useT();
  const [libs, setLibs] = useState<Library[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [editing, setEditing] = useState<string | null>(null); // library id | "new"
  const [openRooms, setOpenRooms] = useState<string | null>(null);

  const load = () =>
    api<Library[]>("/api/libraries")
      .then((l) => { setLibs(l); setFailed(false); })
      .catch(() => setFailed(true));
  useEffect(() => { load(); }, []);
  const saved = () => { setEditing(null); load(); };
  const options = Array.from(new Set([...BASE_AMENITIES, ...(libs ?? []).flatMap((l) => l.amenities)]));

  return (
    <section>
      <div className="mb-2 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold">{t("adminLib.title")}</h2>
        {editing !== "new" && (
          <Button variant="secondary" onClick={() => setEditing("new")}>
            <span aria-hidden>+ </span>{t("adminLib.add")}
          </Button>
        )}
      </div>
      {failed && <ErrorState onRetry={load} />}
      {editing === "new" && (
        <Card className="mb-3">
          <LibraryForm options={options} onSaved={saved} onCancel={() => setEditing(null)} />
        </Card>
      )}
      {libs === null && !failed && <Skeleton className="h-16 w-full" />}
      {libs?.length === 0 && <p className="text-sm text-muted-foreground">{t("adminLib.empty")}</p>}
      {libs && libs.length > 0 && (
        <div className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-surface">
          {libs.map((l) => {
            const name = lang === "th" ? l.name_th : l.name_en;
            return editing === l.id ? (
              <div key={l.id} className="p-4">
                <LibraryForm initial={l} options={options} onSaved={saved} onCancel={() => setEditing(null)} />
              </div>
            ) : (
              <div key={l.id}>
                <div className="flex flex-wrap items-start justify-between gap-3 px-4 py-3">
                  <div className="min-w-0 text-sm">
                    <p className="font-medium">{name}</p>
                    <p className="text-xs text-muted-foreground">
                      {t("adminLib.seats", { n: l.capacity })}
                      {l.amenities.length > 0 && ` · ${l.amenities.join(" · ")}`}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-start gap-2">
                    <Button variant="secondary" aria-expanded={openRooms === l.id} aria-label={`${t("adminLib.rooms")} ${name}`} onClick={() => setOpenRooms(openRooms === l.id ? null : l.id)}>
                      {t("adminLib.rooms")}
                    </Button>
                    <Button variant="secondary" aria-label={`${t("action.edit")} ${name}`} onClick={() => setEditing(l.id)}>
                      {t("action.edit")}
                    </Button>
                    <DeleteButton name={name} onDelete={() => api(`/api/libraries/${l.id}`, { method: "DELETE" }).then(load)} />
                  </div>
                </div>
                {openRooms === l.id && <RoomsPanel library={l} options={options} />}
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
