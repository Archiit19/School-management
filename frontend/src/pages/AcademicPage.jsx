import { useState, useEffect, useCallback } from "react";
import { academicApi } from "../api/client";

export default function AcademicPage() {
  const [tab, setTab] = useState("view");
  const [classes, setClasses] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [classForm, setClassForm] = useState({ name: "", description: "" });
  const [sectionForm, setSectionForm] = useState({ class_id: "", name: "" });
  const [subjectForm, setSubjectForm] = useState({ class_id: "", section_id: "", name: "", code: "" });

  const load = useCallback(async () => {
    try { setClasses(await academicApi.getClasses()); } catch (err) { setError(err.message); }
  }, []);

  useEffect(() => { load(); }, [load]);

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function createClass(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await academicApi.createClass(classForm); msg("Class created."); setClassForm({ name: "", description: "" }); load(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function createSection(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await academicApi.createSection(sectionForm); msg("Section created."); setSectionForm({ class_id: "", name: "" }); load(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function createSubject(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await academicApi.createSubject(subjectForm); msg("Subject created."); setSubjectForm({ class_id: "", section_id: "", name: "", code: "" }); load(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  const flatClasses = classes.map((c) => c.class || c);
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name })));

  const subjectFormSections = subjectForm.class_id ? flatSections.filter((s) => s.class_id === subjectForm.class_id) : [];

  return (
    <>
      <div className="page-header">
        <h1>Academic Structure</h1>
        <p>Flow 3 — Manage classes, sections, and subjects.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "view" ? "active" : ""}`} onClick={() => setTab("view")}>View Tree</button>
        <button className={`tab ${tab === "class" ? "active" : ""}`} onClick={() => setTab("class")}>Add Class</button>
        <button className={`tab ${tab === "section" ? "active" : ""}`} onClick={() => setTab("section")}>Add Section</button>
        <button className={`tab ${tab === "subject" ? "active" : ""}`} onClick={() => setTab("subject")}>Add Subject</button>
      </div>

      {tab === "view" && (
        <div className="card">
          <div className="card-title">Classes / Sections / Subjects <span className="badge badge-get">GET /classes</span></div>
          {classes.length === 0 && <div className="empty">No classes yet. Create one first.</div>}
          {classes.map((c) => {
            const cls = c.class || c;
            return (
              <div key={cls.id} style={{ marginBottom: 16, padding: 12, border: "1px solid var(--clr-border)", borderRadius: "var(--radius)" }}>
                <strong>{cls.name}</strong>{cls.description ? ` — ${cls.description}` : ""}
                <span className="mono text-sm text-muted" style={{ marginLeft: 8 }}>{cls.id}</span>
                {(c.sections || []).length > 0 && (
                  <div style={{ marginTop: 8, marginLeft: 16 }}>
                    <div className="text-sm text-muted" style={{ fontWeight: 600, marginBottom: 4 }}>Sections</div>
                    {c.sections.map((s) => (
                      <div key={s.id} className="text-sm" style={{ padding: "2px 0" }}>
                        {s.name} <span className="mono text-muted">{s.id}</span>
                      </div>
                    ))}
                  </div>
                )}
                {(c.subjects || []).length > 0 && (
                  <div style={{ marginTop: 8, marginLeft: 16 }}>
                    <div className="text-sm text-muted" style={{ fontWeight: 600, marginBottom: 4 }}>Subjects</div>
                    {c.subjects.map((s) => (
                      <div key={s.id} className="text-sm" style={{ padding: "2px 0" }}>
                        {s.name}{s.code ? ` (${s.code})` : ""} <span className="mono text-muted">{s.id}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {tab === "class" && (
        <div className="card">
          <div className="card-title">Create Class <span className="badge badge-post">POST /classes</span></div>
          <form onSubmit={createClass}>
            <div className="grid-2">
              <div className="form-group"><label>Name</label><input required value={classForm.name} onChange={(e) => setClassForm((p) => ({ ...p, name: e.target.value }))} placeholder="Class 10" /></div>
              <div className="form-group"><label>Description</label><input value={classForm.description} onChange={(e) => setClassForm((p) => ({ ...p, description: e.target.value }))} placeholder="Tenth grade" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Class</button></div>
          </form>
        </div>
      )}

      {tab === "section" && (
        <div className="card">
          <div className="card-title">Create Section <span className="badge badge-post">POST /sections</span></div>
          <form onSubmit={createSection}>
            <div className="grid-2">
              <div className="form-group">
                <label>Class</label>
                <select required value={sectionForm.class_id} onChange={(e) => setSectionForm((p) => ({ ...p, class_id: e.target.value }))}>
                  <option value="">Select class...</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group"><label>Section Name</label><input required value={sectionForm.name} onChange={(e) => setSectionForm((p) => ({ ...p, name: e.target.value }))} placeholder="Section A" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Section</button></div>
          </form>
        </div>
      )}

      {tab === "subject" && (
        <div className="card">
          <div className="card-title">Create Subject <span className="badge badge-post">POST /subjects</span></div>
          <form onSubmit={createSubject}>
            <div className="grid-4">
              <div className="form-group">
                <label>Class</label>
                <select required value={subjectForm.class_id} onChange={(e) => setSubjectForm((p) => ({ ...p, class_id: e.target.value, section_id: "" }))}>
                  <option value="">Select class...</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Section (optional)</label>
                <select value={subjectForm.section_id} onChange={(e) => setSubjectForm((p) => ({ ...p, section_id: e.target.value }))} disabled={!subjectForm.class_id}>
                  <option value="">Any section</option>
                  {subjectFormSections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="form-group"><label>Subject Name</label><input required value={subjectForm.name} onChange={(e) => setSubjectForm((p) => ({ ...p, name: e.target.value }))} placeholder="Mathematics" /></div>
              <div className="form-group"><label>Code</label><input value={subjectForm.code} onChange={(e) => setSubjectForm((p) => ({ ...p, code: e.target.value }))} placeholder="MATH101" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Subject</button></div>
          </form>
        </div>
      )}
    </>
  );
}
