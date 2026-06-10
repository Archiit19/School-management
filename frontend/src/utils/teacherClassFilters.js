/** Filter sections a teacher may use for a class (from teacher_assignments + subject section links). */
export function sectionsForAssignedClass(classNode, classId, teacherAssignments, isTeacher) {
  if (!classNode) return [];
  const sections = classNode.sections || [];
  if (!isTeacher) return sections;
  if (sections.length === 0) return [];

  const subjects = classNode.subjects || [];
  const assignedIds = new Set(
    teacherAssignments.filter((ta) => ta.class_id === classId).map((ta) => ta.subject_id),
  );
  const hasClassWide = subjects.some((s) => assignedIds.has(s.id) && !s.section_id);
  if (hasClassWide) return sections;

  const sectionIds = new Set();
  subjects.forEach((s) => {
    if (assignedIds.has(s.id) && s.section_id) sectionIds.add(s.section_id);
  });
  return sections.filter((sec) => sectionIds.has(sec.id));
}

/** Subjects for a class + section; teachers only see subjects they are assigned to teach. */
export function subjectsForClassSection(classNode, classId, sectionId, teacherAssignments, isTeacher) {
  if (!classNode) return [];
  const subjects = classNode.subjects || [];
  const sections = classNode.sections || [];
  const hasSections = sections.length > 0;

  if (!isTeacher) {
    if (!hasSections) return subjects.filter((s) => !s.section_id);
    if (!sectionId) return [];
    return subjects.filter((s) => !s.section_id || s.section_id === sectionId);
  }

  const assignedIds = new Set(
    teacherAssignments.filter((ta) => ta.class_id === classId).map((ta) => ta.subject_id),
  );
  return subjects.filter((s) => {
    if (!assignedIds.has(s.id)) return false;
    if (hasSections) {
      if (!sectionId) return false;
      if (!s.section_id) return true;
      return s.section_id === sectionId;
    }
    return !s.section_id;
  });
}

/** Classes where the teacher has at least one teacher_assignment row. */
export function classesForTeacher(allClassNodes, teacherAssignments, isTeacher) {
  const allClasses = allClassNodes.map((c) => c.class || c);
  if (!isTeacher) return allClasses;
  return allClasses.filter((c) => teacherAssignments.some((ta) => ta.class_id === c.id));
}
