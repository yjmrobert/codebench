import { describe, it, expect } from 'vitest';
import { toLetterGrade, toSubGrade } from '../grading.js';

describe('toLetterGrade', () => {
  it('returns A for scores >= 90', () => {
    expect(toLetterGrade(90)).toBe('A');
    expect(toLetterGrade(100)).toBe('A');
    expect(toLetterGrade(95)).toBe('A');
  });

  it('returns B for scores 80-89', () => {
    expect(toLetterGrade(80)).toBe('B');
    expect(toLetterGrade(89)).toBe('B');
  });

  it('returns C for scores 70-79', () => {
    expect(toLetterGrade(70)).toBe('C');
    expect(toLetterGrade(79)).toBe('C');
  });

  it('returns D for scores 60-69', () => {
    expect(toLetterGrade(60)).toBe('D');
    expect(toLetterGrade(69)).toBe('D');
  });

  it('returns F for scores < 60', () => {
    expect(toLetterGrade(59)).toBe('F');
    expect(toLetterGrade(0)).toBe('F');
  });
});

describe('toSubGrade', () => {
  it('returns A+ for 97+', () => {
    expect(toSubGrade(97)).toBe('A+');
  });

  it('returns A for 93-96', () => {
    expect(toSubGrade(95)).toBe('A');
  });

  it('returns A- for 90-92', () => {
    expect(toSubGrade(90)).toBe('A-');
  });

  it('returns F for scores < 60', () => {
    expect(toSubGrade(50)).toBe('F');
  });
});
