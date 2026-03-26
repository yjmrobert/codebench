#!/usr/bin/env node

import { Command } from 'commander';
import { runAnalysis } from './commands/analyze.js';
import { runInit } from './commands/init.js';
import { runHistory } from './commands/history.js';

const program = new Command();

program
  .name('codebench')
  .description('Analyze codebases and produce a unified health score')
  .version('0.1.0');

// Default command: full analysis
program
  .argument('[path]', 'Path to analyze', process.cwd())
  .option('--json', 'Output as JSON')
  .option('--format <fmt>', 'Output format: terminal | json | markdown')
  .option('--metric <name>', 'Run a single metric')
  .option('--threshold <score>', 'Fail if score is below threshold', parseInt)
  .option('--config <path>', 'Path to config file')
  .option('--no-cache', 'Skip cached results')
  .action(async (path: string, options) => {
    const cwd = path || process.cwd();
    const exitCode = await runAnalysis(cwd, options);
    process.exit(exitCode);
  });

// Init command
program
  .command('init')
  .description('Generate a .codebench.yml config file')
  .action(async () => {
    const exitCode = await runInit(process.cwd());
    process.exit(exitCode);
  });

// History command
program
  .command('history')
  .description('Show score trend over time')
  .option('--limit <n>', 'Number of entries to show', parseInt)
  .action(async (options) => {
    const exitCode = await runHistory(process.cwd(), options);
    process.exit(exitCode);
  });

program.parse();
