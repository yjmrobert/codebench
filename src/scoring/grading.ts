import type { LetterGrade } from '../config/index.js';

export function toLetterGrade(score: number): LetterGrade {
  if (score >= 90) return 'A';
  if (score >= 80) return 'B';
  if (score >= 70) return 'C';
  if (score >= 60) return 'D';
  return 'F';
}

export function toSubGrade(score: number): string {
  const grade = toLetterGrade(score);
  if (grade === 'F') return 'F';
  const withinBand = score % 10;
  if (withinBand >= 7) return `${grade}+`;
  if (withinBand >= 3) return grade;
  return `${grade}-`;
}
